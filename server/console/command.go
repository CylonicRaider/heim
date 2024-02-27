package console

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode"
)

type BinaryValue []byte

func (bv *BinaryValue) String() string {
	if bv == nil || len(*bv) == 0 {
		return ""
	}
	return "\\x" + strings.ToUpper(hex.EncodeToString(*bv))
}

func (bv *BinaryValue) Get() interface{} {
	return []byte(*bv)
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

type StringSliceValue []string

func (ssv *StringSliceValue) String() string {
	if ssv == nil || len(*ssv) == 0 {
		return ""
	}
	return strings.Join(*ssv, ", ")
}

func (ssv *StringSliceValue) Get() interface{} {
	return []string(*ssv)
}

func (ssv *StringSliceValue) Set(value string) error {
	*ssv = append(*ssv, value)
	return nil
}

type DurationValue struct {
	time.Duration
}

func (dv *DurationValue) String() string {
	return dv.Duration.String()
}

func (dv *DurationValue) Get() interface{} {
	return dv.Duration
}

func (dv *DurationValue) Set(value string) error {
	d, err := time.ParseDuration(value)
	if err != nil {
		return err
	}
	dv.Duration = d
	return nil
}

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

type flags struct {
	name         string
	flags        *flag.FlagSet
	args         *flag.FlagSet
	reqFlags     map[string]bool
	argsOrder    []string
	firstOptArg  int
	restArgIndex int
}

func newFlags(name string) *flags {
	result := &flags{
		name:         name,
		flags:        flag.NewFlagSet(name, flag.ContinueOnError),
		args:         flag.NewFlagSet(name, flag.PanicOnError),
		reqFlags:     nil,
		argsOrder:    []string{},
		firstOptArg:  0,
		restArgIndex: -1,
	}
	result.flags.Usage = result.Usage
	result.args.Usage = result.Usage
	return result
}

func (f *flags) VisitAll(cb func(*flag.Flag, bool, bool, bool)) {
	f.flags.VisitAll(func(fl *flag.Flag) { cb(fl, false, f.reqFlags[fl.Name], false) })
	for i, name := range f.argsOrder {
		cb(f.args.Lookup(name), true, (i < f.firstOptArg), (i == f.restArgIndex))
	}
}

func (f *flags) Usage() {
	buf := &bytes.Buffer{}
	name := f.name
	if name == "" {
		name = "..."
	}
	fmt.Fprintf(buf, "USAGE: %s", name)

	f.VisitAll(func(fl *flag.Flag, positional, required, rest bool) {
		buf.WriteByte(' ')
		if !required {
			buf.WriteByte('[')
		}
		if positional {
			buf.WriteString(fl.Name)
			if rest {
				buf.WriteString(" ...")
			}
		} else {
			valName, _ := flag.UnquoteUsage(fl)
			buf.WriteByte('-')
			buf.WriteString(fl.Name)
			if valName != "" {
				buf.WriteByte(' ')
				buf.WriteString(valName)
			}
		}
		if !required {
			buf.WriteByte(']')
		}
	})

	fmt.Fprintln(f.flags.Output(), buf.String())
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
		if argIndex != f.restArgIndex {
			argIndex++
		}
	}

	seenFlags := map[string]bool{}
	f.flags.Visit(func(fl *flag.Flag) { seenFlags[fl.Name] = true })
	for name, required := range f.reqFlags {
		if required && !seenFlags[name] {
			return f.failf("missing value for required flag -%s",
				name)
		}
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
			// Required arguments' defaults are of little use.
		} else if text, show := defaultValueString(argFlag); show {
			fmt.Fprintf(buf, " (default %s)", text)
		}
		fmt.Fprintln(buf)
	}
	fmt.Fprint(f.args.Output(), buf.String())
}

type CLIEnv interface {
	Console
}

type CommandParams interface {
	Run(env CLIEnv) error
}

func cloneParams(template CommandParams) (reflect.Type, reflect.Value, reflect.Value) {
	defaults := reflect.ValueOf(template)
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

	return tp, defaults, params
}

func CloneParams(p CommandParams) CommandParams {
	_, _, result := cloneParams(p)
	return result.Interface().(CommandParams)
}

type Commander interface {
	GetName() string
	GetDesc() string
	GetDefaults() CommandParams
	Run(env CLIEnv, args []string) error
	Help(con Console)
}

type Command struct {
	Name     string
	Desc     string
	Defaults CommandParams
}

func (c *Command) GetName() string {
	return c.Name
}

func (c *Command) GetDesc() string {
	return c.Desc
}

func (c *Command) GetDefaults() CommandParams {
	return c.Defaults
}

func (c *Command) flags() (*flags, CommandParams) {
	flagGetterType := reflect.TypeOf((*flag.Getter)(nil)).Elem()
	unknownType := func(field reflect.StructField) {
		panic("unsupported command parameter type: " + field.Type.String())
	}

	result := newFlags(c.Name)
	tp, defaults, params := cloneParams(c.Defaults)

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
			if result.restArgIndex != -1 {
				panic("cannot have positional argument after rest argument")
			}
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
			if result.reqFlags == nil {
				result.reqFlags = map[string]bool{fd.name: true}
			} else {
				result.reqFlags[fd.name] = true
			}
		}

		isRest := false
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
			case reflect.String:
				flags.Var((*StringSliceValue)(vp.(*[]string)), fd.name, fd.usage)
				isRest = true
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

		if fd.positional && isRest {
			result.restArgIndex = len(result.argsOrder) - 1
		}
	}

	return result, params.Addr().Interface().(CommandParams)
}

func (c *Command) Parse(con Console, argv []string) CommandParams {
	flags, params := c.flags()
	flags.SetOutput(con)
	err := flags.Parse(argv)
	if err == flag.ErrHelp {
		// flags has already written the usage, append the help
		c.helpDetail(con, flags, params)
		return nil
	} else if err != nil {
		// flags has already written an error message
		return nil
	}
	return params
}

func (c *Command) Run(env CLIEnv, argv []string) error {
	params := c.Parse(env, argv)
	if params == nil {
		return nil
	}
	return params.Run(env)
}

func (c *Command) Help(con Console) {
	flags, params := c.flags()
	flags.SetOutput(con)
	flags.Usage()
	c.helpDetail(con, flags, params)
}

func (c *Command) helpDetail(con Console, flags *flags, params CommandParams) {
	if flags == nil {
		flags, _ := c.flags()
		flags.SetOutput(con)
	}
	if c.Desc != "" {
		con.Println(c.Desc)
	}
	flags.PrintDefaults()
	if np, ok := params.(NestedCommandParams); ok {
		np.Commands().help(con, false, true)
	}
}

type HiddenCommander interface {
	Commander
	IsHidden() bool
}

type HiddenCommand struct {
	Command
}

func (c *HiddenCommand) IsHidden() bool {
	return true
}
