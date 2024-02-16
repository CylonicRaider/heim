package main

import (
	"flag"
	"fmt"
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

type CommandParams interface {
	Run(con Console, argv []string)
}

type Command struct {
	Name     string
	Defaults CommandParams
}

type fieldDesc struct {
	name  string
	usage string
}

func newFieldDesc(field reflect.StructField) *fieldDesc {
	tags := field.Tag
	usage, ok := tags.Lookup("usage")
	if !ok {
		return nil
	}
	name, ok := tags.Lookup("cli")
	if !ok {
		for i, c := range field.Name {
			if unicode.IsUpper(c) && i != 0 {
				name += "-"
			}
			name += string(unicode.ToLower(c))
		}
	}
	return &fieldDesc{
		name:  name,
		usage: usage,
	}
}

func (c *Command) Flags() (*flag.FlagSet, CommandParams) {
	flags := flag.NewFlagSet(c.Name, flag.ContinueOnError)

	defaults := reflect.ValueOf(c.Defaults)
	tp := defaults.Type()
	if tp.Kind() != reflect.Pointer {
		panic("Command parameters must be pointer to struct, got " +
			tp.Kind().String())
	}
	defaults = defaults.Elem()
	tp = defaults.Type()
	if tp.Kind() != reflect.Struct {
		panic("Command parameters must be pointer to struct, got pointer to " +
			tp.Kind().String())
	}

	params := reflect.New(tp).Elem()
	params.Set(defaults)

	for _, field := range reflect.VisibleFields(tp) {
		if field.Anonymous || field.PkgPath != "" {
			continue
		}

		ov := defaults.FieldByIndex(field.Index).Interface()
		vp := params.FieldByIndex(field.Index).Addr().Interface()
		fd := newFieldDesc(field)
		if fd == nil {
			continue
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
		default:
			panic("Unsupported command parameter type: " + field.Type.Kind().String())
		}
	}

	return flags, params.Addr().Interface().(CommandParams)
}

func (c *Command) Parse(con Console, argv []string) (CommandParams, []string) {
	flags, params := c.Flags()
	flags.SetOutput(con)
	err := flags.Parse(argv)
	if err == flag.ErrHelp {
		return nil, nil
	} else if err != nil {
		con.Println("ERROR: " + err.Error())
		return nil, nil
	}
	return params, flags.Args()
}

func (c *Command) Run(con Console, argv []string) {
	params, restArgv := c.Parse(con, argv)
	if params == nil {
		return
	}
	params.Run(con, restArgv)
}

func (c *Command) Help(con Console) {
	con.Println("Usage of " + c.Name + ":")
	flags, _ := c.Flags()
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
		} else if words[0] == "quit" {
			break
		} else if words[0] == "help" {
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
