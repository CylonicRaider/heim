package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

type Console interface {
	io.Writer
	io.Closer

	Print(text string) error
	Println(text string) error
	Printf(format string, args ...interface{}) error

	ReadLine(prompt string) (string, error)
	ReadPass(prompt string) (string, error)
}

type DefaultConsole struct {
	term  *term.Terminal
	fd    int
	state *term.State
}

func NewDefaultConsole() Console {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	panicIfFailed(err)
	return &DefaultConsole{
		term:  term.NewTerminal(os.Stdin, ""),
		fd:    fd,
		state: oldState,
	}
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
	err := term.Restore(c.fd, c.state)
	panicIfFailed(err)
	return nil
}

func (c *DefaultConsole) Print(text string) error {
	_, err := c.Write([]byte(text))
	return err
}

func (c *DefaultConsole) Println(text string) error {
	return c.Print(text + "\n")
}

func (c *DefaultConsole) Printf(format string, args ...interface{}) error {
	_, err := fmt.Fprintf(c, format, args...)
	return err
}

func (c *DefaultConsole) ReadLine(prompt string) (string, error) {
	c.term.SetPrompt(prompt)
	return c.term.ReadLine()
}

func (c *DefaultConsole) ReadPass(prompt string) (string, error) {
	return c.term.ReadPassword(prompt)
}
