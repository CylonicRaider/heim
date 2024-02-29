package console

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var (
	spaceRe = regexp.MustCompile("^\\s+")
	wordRe  = regexp.MustCompile("^([^\"\\s]+|\"([^\"\\\\]+|\\\\.)*\")")
)

type CommandSet map[string]Commander

func (cs CommandSet) Add(cmd Commander) { cs[cmd.GetName()] = cmd }

func (cs CommandSet) GetCommand(con Console, name string) Commander {
	cmd, ok := cs[name]
	if !ok {
		con.Println("unknown command " + name)
		return nil
	}
	return cmd
}

func (cs CommandSet) Run(env CLIEnv, argv []string) error {
	if len(argv) == 0 {
		return nil
	}
	cmd := cs.GetCommand(env, argv[0])
	if cmd == nil {
		return nil
	}
	return cmd.Run(env, argv[1:])
}

func (cs CommandSet) help(con Console, full bool, saySub bool) {
	allCommands := []Commander{}
	for _, cmd := range cs {
		if !full {
			if hc, ok := cmd.(HiddenCommander); ok && hc.IsHidden() {
				continue
			}
		}
		allCommands = append(allCommands, cmd)
	}
	sort.Slice(allCommands, func(i, j int) bool {
		return allCommands[i].GetName() < allCommands[j].GetName()
	})

	buf := &bytes.Buffer{}
	if saySub {
		buf.WriteString("Known subcommands:\n")
	} else {
		buf.WriteString("Known commands:\n")
	}
	for _, cmd := range allCommands {
		name, desc := cmd.GetName(), cmd.GetDesc()
		nameLen := utf8.RuneCountInString(name)
		fmt.Fprintf(buf, "  %s", name)
		if nameLen < 6 {
			buf.WriteString(strings.Repeat(" ", 6-nameLen))
		} else {
			buf.WriteString("\n        ")
		}
		fmt.Fprintln(buf, strings.ReplaceAll(desc, "\n", "\n        "))
	}
	con.Print(buf.String())
}

func (cs CommandSet) Help(con Console) {
	cs.help(con, false, false)
}

type HelpCmd struct {
	Full    bool   `usage:"Also list hidden commands."`
	Command string `usage:"Command to get help of." cli:",arg"`
}

func (c *HelpCmd) Run(env CLIEnv) error {
	return HelpError(*c)
}

type HelpError HelpCmd

func (he HelpError) Error() string {
	if he.Command != "" {
		return "user requested help of " + he.Command
	} else if he.Full {
		return "user requested full command list"
	} else {
		return "user requested command list"
	}
}

type ExitCmd struct{}

func (c *ExitCmd) Run(env CLIEnv) error {
	return ExitError(*c)
}

type ExitError ExitCmd

func (ee ExitError) Error() string {
	return "user requested exit"
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
		m := spaceRe.FindStringIndex(line[offset:])
		if m == nil || m[0] != 0 {
			if !first {
				return fail("unexpected characters")
			}
		} else {
			offset += m[1]
		}
		if offset == len(line) {
			break
		}
		m = wordRe.FindStringIndex(line[offset:])
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

type NestedCommandParams interface {
	CommandParams
	Commands() CommandSet
}

type CLICallback func(cli *CLI, done bool, argv []string, err error) error

type CLI struct {
	Console
	Prompt   string
	Argv     []string `usage:"Single command to run." cli:"cmd,arg"`
	Callback CLICallback
	commands CommandSet
}

func NewCLI(prompt string, standardCommands bool, commands ...*Command) *CLI {
	result := &CLI{Prompt: prompt, commands: CommandSet{}}
	if standardCommands {
		result.AddStandardCommands()
	}
	for _, cmd := range commands {
		result.commands.Add(cmd)
	}
	return result
}

func (c *CLI) Commands() CommandSet {
	return c.commands
}

func (c *CLI) AddStandardCommands() {
	c.addNewHiddenCommand("help", "Print the usage details of commands.", &HelpCmd{})
	c.addNewHiddenCommand("exit", "Exit the command line.", &ExitCmd{})
	c.addNewHiddenCommand("quit", "Exit the command line.", &ExitCmd{})
}

func (c *CLI) AddCommand(cmd Commander) {
	c.commands.Add(cmd)
}

func (c *CLI) AddNewCommand(name, desc string, defaults CommandParams) {
	c.AddCommand(&Command{Name: name, Desc: desc, Defaults: defaults})
}

func (c *CLI) addNewHiddenCommand(name, desc string, defaults CommandParams) {
	c.AddCommand(&HiddenCommand{Command{Name: name, Desc: desc, Defaults: defaults}})
}

func (c *CLI) runOne(argv []string) error {
	var err error
	if c.Callback != nil {
		err = c.Callback(c, false, argv, nil)
	}
	if err == nil {
		err = c.commands.Run(c, argv)
		if c.Callback != nil {
			err = c.Callback(c, true, argv, err)
		}
	}
	if help, ok := err.(HelpError); ok {
		c.runHelp(help)
		err = nil
	}
	return err
}

func (c *CLI) runHelp(help HelpError) {
	if help.Command == "" {
		c.commands.help(c, help.Full, false)
		return
	}
	cmd := c.commands.GetCommand(c, help.Command)
	if cmd == nil {
		return
	}
	cmd.Help(c)
}

func (c *CLI) runLoop() error {
	for {
		line, err := c.ReadLine(c.Prompt)
		if err != nil {
			if err != io.EOF {
				c.Println("ERROR:", err)
			}
			break
		}

		argv, err := SplitLine(line)
		if err != nil {
			c.Println(err)
			continue
		}

		err = c.runOne(argv)
		if _, ok := err.(ExitError); ok {
			break
		} else if err != nil && err != io.EOF {
			c.Println("ERROR:", err)
		}
	}
	return nil
}

func (c *CLI) Run(parent CLIEnv) error {
	if c.Console != nil {
		panic("trying to run CLI already bound to console")
	}
	c.Console = parent

	if len(c.Argv) == 0 {
		return c.runLoop()
	} else {
		return c.runOne(c.Argv)
	}
}
