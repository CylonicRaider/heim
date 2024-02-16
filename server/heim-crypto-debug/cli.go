package main

import (
	"flag"
	"reflect"
	"unicode"
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

type CommandSet map[string]*Command

func NewCommandSet() CommandSet { return CommandSet{} }

func (cs CommandSet) Add(cmd *Command) { cs[cmd.Name] = cmd }

func (cs CommandSet) AddNew(name string, defaults CommandParams) {
	cs.Add(&Command{name, defaults})
}

func (cs CommandSet) Run(con Console, argv []string) {
	if len(argv) == 0 {
		return
	}
	cmd, ok := cs[argv[0]]
	if !ok {
		con.Println("ERROR: unknown command " + argv[0])
		return
	}
	cmd.Run(con, argv[1:])
}
