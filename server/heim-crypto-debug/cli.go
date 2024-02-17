package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	spaceRe = regexp.MustCompile("^\\s+")
	wordRe  = regexp.MustCompile("^([^\"\\s]+|\"([^\"\\\\]+|\\\\.)*\")")
)

type fieldDesc struct {
	name       string
	usage      string
	positional bool
	required   bool
}

func newFieldDesc(field reflect.StructField) *fieldDesc {
	tags := field.Tag
	usage, ok := tags.Lookup("usage")
	if !ok {
		return nil
	}
	info, _ := tags.Lookup("cli")
	infoParts := strings.Split(info, ",")
	name := infoParts[0]
	if name == "" {
		for i, c := range field.Name {
			if unicode.IsUpper(c) && i != 0 {
				name += "-"
			}
			name += string(unicode.ToLower(c))
		}
	}
	result := &fieldDesc{
		name:  name,
		usage: usage,
	}
	for _, tag := range infoParts[1:] {
		switch tag {
		case "arg":
			result.positional = true
		case "required":
			result.required = true
		default:
			panic(fmt.Sprintf("unrecognized CLI tag: %s", tag))
		}
	}
	return result
}

type BinaryValue []byte

func (bv *BinaryValue) String() string {
	if bv == nil || len(*bv) == 0 {
		return "\"\""
	}
	return "\\x" + strings.ToUpper(hex.EncodeToString(*bv))
}

func (bv *BinaryValue) Get() interface{} {
	return *bv
}

func (bv *BinaryValue) Set(value string) error {
	if value == "" || value[0] != '\\' {
		*bv = []byte(value)
		return nil
	} else if value[1] != 'x' {
		return fmt.Errorf("invalid byte string literal: expected \\x")
	}
	decoded, err := hex.DecodeString(value[2:])
	if err != nil {
		return err
	}
	*bv = decoded
	return nil
}

type flags struct {
	flags       *flag.FlagSet
	args        *flag.FlagSet
	argsOrder   []string
	firstOptArg int
}

func newFlags(name string) *flags {
	result := &flags{
		flags:       flag.NewFlagSet(name, flag.ContinueOnError),
		args:        flag.NewFlagSet(name, flag.PanicOnError),
		argsOrder:   []string{},
		firstOptArg: 0,
	}
	result.flags.Usage = result.Usage
	result.args.Usage = result.Usage
	return result
}

func (f *flags) Usage() {
	fmt.Fprintf(f.flags.Output(), "See help %s for usage.\n", f.flags.Name())
}

func (f *flags) SetOutput(w io.Writer) {
	f.flags.SetOutput(w)
	f.args.SetOutput(w)
}

func (f *flags) failf(format string, values ...interface{}) error {
	err := fmt.Errorf(format, values...)
	fmt.Fprintln(f.flags.Output(), err)
	f.Usage()
	return err
}

func (f *flags) Parse(argv []string) error {
	if err := f.flags.Parse(argv); err != nil {
		return err
	}
	argv = f.flags.Args()

	argIndex := 0
	for _, arg := range argv {
		if argIndex >= len(f.argsOrder) {
			return f.failf("too many positional arguments")
		}
		argName := f.argsOrder[argIndex]
		if err := f.args.Set(argName, arg); err != nil {
			return f.failf("invalid value %q or argument %s: %v",
				arg, argName, err)
		}
		argIndex++
	}
	if argIndex < f.firstOptArg {
		return f.failf("missing value for required argument %s",
			f.argsOrder[argIndex])
	}
	return nil
}

func defaultValueString(f *flag.Flag) (string, bool) {
	zv := reflect.New(reflect.TypeOf(f.Value).Elem())
	zvs := zv.Interface().(flag.Value).String()
	if f.DefValue == zvs {
		return "", false
	}

	if zv.Kind() == reflect.String {
		return fmt.Sprintf("%q", f.DefValue), true
	} else {
		return f.DefValue, true
	}
}

// Yes, that name is stupid. No, I did not invent it.
func (f *flags) PrintDefaults() {
	f.flags.PrintDefaults()

	buf := &bytes.Buffer{}
	for i, name := range f.argsOrder {
		argFlag := f.args.Lookup(name)
		valName, usage := flag.UnquoteUsage(argFlag)
		fmt.Fprintf(buf, "  %s", name)
		if valName != "" && valName != name {
			fmt.Fprintf(buf, ": %s", valName)
		}
		fmt.Fprintf(buf, "\n    \t%s",
			strings.ReplaceAll(usage, "\n", "\n    \t"))
		if i < f.firstOptArg {
			fmt.Fprintf(buf, " (required)")
		} else if text, show := defaultValueString(argFlag); show {
			fmt.Fprintf(buf, " (default %s)", text)
		}
		fmt.Fprintln(buf)
	}
	fmt.Fprint(f.args.Output(), buf.String())
}

type CommandParams interface {
	Run(con Console)
}

type Command struct {
	Name     string
	Defaults CommandParams
}

func (c *Command) flags() (*flags, CommandParams) {
	flagGetterType := reflect.TypeOf((*flag.Getter)(nil)).Elem()
	unknownType := func(field reflect.StructField) {
		panic("unsupported command parameter type: " + field.Type.String())
	}

	result := newFlags(c.Name)

	defaults := reflect.ValueOf(c.Defaults)
	tp := defaults.Type()
	if tp.Kind() != reflect.Pointer {
		panic("command parameters must be pointer to struct, got " +
			tp.Kind().String())
	}
	defaults = defaults.Elem()
	tp = defaults.Type()
	if tp.Kind() != reflect.Struct {
		panic("command parameters must be pointer to struct, got pointer to " +
			tp.Kind().String())
	}

	params := reflect.New(tp).Elem()
	params.Set(defaults)

	positionalRequired := true
	for _, field := range reflect.VisibleFields(tp) {
		if field.Anonymous || field.PkgPath != "" {
			continue
		}

		fd := newFieldDesc(field)
		if fd == nil {
			continue
		}

		ov := defaults.FieldByIndex(field.Index).Interface()
		vp := params.FieldByIndex(field.Index).Addr().Interface()
		flags := result.flags
		if fd.positional {
			result.argsOrder = append(result.argsOrder, fd.name)
			if fd.required {
				if !positionalRequired {
					panic("required positional argument cannot follow optional one")
				}
				result.firstOptArg = len(result.argsOrder)
			} else {
				positionalRequired = false
			}
			flags = result.args
		} else if fd.required {
			panic("only positional arguments can be required")
		}

		switch field.Type.Kind() {
		case reflect.Bool:
			flags.BoolVar(vp.(*bool), fd.name, ov.(bool), fd.usage)
		case reflect.Int:
			flags.IntVar(vp.(*int), fd.name, ov.(int), fd.usage)
		case reflect.Int64:
			flags.Int64Var(vp.(*int64), fd.name, ov.(int64), fd.usage)
		case reflect.Uint:
			flags.UintVar(vp.(*uint), fd.name, ov.(uint), fd.usage)
		case reflect.Uint64:
			flags.Uint64Var(vp.(*uint64), fd.name, ov.(uint64), fd.usage)
		case reflect.Float64:
			flags.Float64Var(vp.(*float64), fd.name, ov.(float64), fd.usage)
		case reflect.String:
			flags.StringVar(vp.(*string), fd.name, ov.(string), fd.usage)
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				flags.Var((*BinaryValue)(vp.(*[]byte)), fd.name, fd.usage)
			default:
				unknownType(field)
			}
		case reflect.Struct:
			if reflect.PointerTo(field.Type).Implements(flagGetterType) {
				flags.Var(vp.(flag.Getter), fd.name, fd.usage)
			} else {
				unknownType(field)
			}
		default:
			unknownType(field)
		}
	}

	return result, params.Addr().Interface().(CommandParams)
}

func (c *Command) Parse(con Console, argv []string) CommandParams {
	flags, params := c.flags()
	flags.SetOutput(con)
	err := flags.Parse(argv)
	if err != nil {
		// flags has already written an error message
		return nil
	}
	return params
}

func (c *Command) Run(con Console, argv []string) {
	params := c.Parse(con, argv)
	if params == nil {
		return
	}
	params.Run(con)
}

func (c *Command) Help(con Console) {
	con.Println("Usage of " + c.Name + ":")
	flags, _ := c.flags()
	flags.SetOutput(con)
	flags.PrintDefaults()
}

type CommandSet map[string]*Command

func NewCommandSet() CommandSet { return CommandSet{} }

func (cs CommandSet) Add(cmd *Command) { cs[cmd.Name] = cmd }

func (cs CommandSet) AddNew(name string, defaults CommandParams) {
	cs.Add(&Command{name, defaults})
}

func (cs CommandSet) GetCommand(con Console, name string) *Command {
	cmd, ok := cs[name]
	if !ok {
		con.Println("ERROR: unknown command " + name)
		return nil
	}
	return cmd
}

func (cs CommandSet) Run(con Console, argv []string) {
	if len(argv) == 0 {
		return
	}
	cmd := cs.GetCommand(con, argv[0])
	if cmd == nil {
		return
	}
	cmd.Run(con, argv[1:])
}

func parseWord(word string) string {
	if word[0] != '"' {
		return word
	}

	result := []byte{}
	for offset := 1; offset < len(word)-1; {
		shift := strings.IndexByte(word[offset:], '\\')
		if shift == -1 {
			result = append(result, word[offset:len(word)-1]...)
			break
		}
		result = append(result, word[offset:offset+shift]...)
		offset += shift

		nextOffset := offset + 2
		if word[offset+1] == '\\' || word[offset+1] == '"' {
			offset++
		}
		result = append(result, word[offset:nextOffset]...)
		offset = nextOffset
	}
	return string(result)
}

func SplitLine(line string) ([]string, error) {
	offset := 0
	fail := func(msg string) ([]string, error) {
		return nil, fmt.Errorf("offset %d: %s", offset, msg)
	}
	result := []string{}
	for first := true; offset != len(line); first = false {
		if !first {
			m := spaceRe.FindStringIndex(line[offset:])
			if m == nil || m[0] != 0 {
				return fail("unexpected characters")
			}
			offset += m[1]
		}
		m := wordRe.FindStringIndex(line[offset:])
		if m == nil {
			return fail("syntax error")
		} else if m[0] != 0 {
			return fail("unexpected characters")
		}
		result = append(result, parseWord(line[offset:offset+m[1]]))
		offset += m[1]
	}
	return result, nil
}

type CLI struct {
	Prompt   string
	Commands CommandSet
}

func NewCLI(prompt string) *CLI {
	return &CLI{Prompt: prompt, Commands: CommandSet{}}
}

func (c *CLI) Help(con Console, cmdName *string) {
	if cmdName == nil {
		allNames := []string{"help", "quit"}
		for name, _ := range c.Commands {
			allNames = append(allNames, name)
		}
		sort.Strings(allNames)
		msg := []byte("Known commands: ")
		for i, name := range allNames {
			if i != 0 {
				msg = append(msg, ", "...)
			}
			msg = append(msg, name...)
		}
		con.Println(string(msg))
		return
	} else if *cmdName == "help" {
		con.Println("Print a list of commands or the help of a particular command")
		return
	} else if *cmdName == "quit" {
		con.Println("Exit")
		return
	}

	cmd := c.Commands.GetCommand(con, *cmdName)
	if cmd == nil {
		return
	}
	cmd.Help(con)
}

func (c *CLI) Run(con Console) {
	for {
		line := con.ReadLine(c.Prompt)
		if line == nil {
			break
		}
		words, err := SplitLine(*line)
		if err != nil {
			con.Println("ERROR: " + err.Error())
			continue
		} else if len(words) > 0 && words[0] == "quit" {
			break
		} else if len(words) > 0 && words[0] == "help" {
			if len(words) == 1 {
				c.Help(con, nil)
			} else {
				c.Help(con, &words[1])
			}
		} else {
			c.Commands.Run(con, words)
		}
	}
}
