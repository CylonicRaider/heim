package console

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type ReadWriteFDer interface {
	io.ReadWriter
	Fd() uintptr
}

type Console interface {
	io.Writer
	io.Closer

	Print(args ...interface{}) error
	Println(args ...interface{}) error
	Printf(format string, args ...interface{}) error

	ReadLine(prompt string) (string, error)
	ReadPassword(prompt string) (string, error)
}

type DefaultConsole struct {
	term  *term.Terminal
	fd    int
	state *term.State
}

func NewDefaultConsole(s io.ReadWriter) *DefaultConsole {
	return &DefaultConsole{
		term: term.NewTerminal(s, ""),
		fd:   -1,
	}
}

func NewRawDefaultConsole(s ReadWriteFDer) (*DefaultConsole, error) {
	fd := int(s.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	return &DefaultConsole{
		term:  term.NewTerminal(s, ""),
		fd:    fd,
		state: oldState,
	}, nil
}

func NewStdioConsole() *DefaultConsole {
	result, err := NewRawDefaultConsole(os.Stdin)
	if err != nil {
		panic(err)
	}
	return result
}

func (c *DefaultConsole) Write(data []byte) (n int, err error) {
	written, err := c.term.Write(data)
	if written != len(data) {
		panic(fmt.Errorf("Incomplete write to console?! Passed %d, wrote %d",
			len(data), written))
	}
	return written, err
}

func (c *DefaultConsole) Close() error {
	if c.fd == -1 {
		return nil
	}
	return term.Restore(c.fd, c.state)
}

func (c *DefaultConsole) Print(args ...interface{}) error {
	_, err := fmt.Fprint(c, args...)
	return err
}

func (c *DefaultConsole) Println(args ...interface{}) error {
	_, err := fmt.Fprintln(c, args...)
	return err
}

func (c *DefaultConsole) Printf(format string, args ...interface{}) error {
	_, err := fmt.Fprintf(c, format, args...)
	return err
}

func (c *DefaultConsole) ReadLine(prompt string) (string, error) {
	c.term.SetPrompt(prompt)
	return c.term.ReadLine()
}

func (c *DefaultConsole) ReadPassword(prompt string) (string, error) {
	return c.term.ReadPassword(prompt)
}
