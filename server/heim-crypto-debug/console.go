package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

func panicIfFailed(err error) {
	if err != nil {
		panic(err)
	}
}

type Console interface {
	io.Closer

	Print(text string)
	Println(text string)

	ReadLine(prompt string) string
	ReadPass(prompt string) string
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

func (c *DefaultConsole) Close() error {
	err := term.Restore(c.fd, c.state)
	panicIfFailed(err)
	return nil
}

func (c *DefaultConsole) Print(text string) {
	written, err := c.term.Write([]byte(text))
	panicIfFailed(err)
	if written != len(text) {
		panic(fmt.Errorf("Incomplete write to console?! passed %d, wrote %d",
			len(text), written))
	}
}

func (c *DefaultConsole) Println(text string) {
	c.Print(text + "\n")
}

func (c *DefaultConsole) ReadLine(prompt string) string {
	c.term.SetPrompt(prompt)
	res, err := c.term.ReadLine()
	panicIfFailed(err)
	return res
}

func (c *DefaultConsole) ReadPass(prompt string) string {
	res, err := c.term.ReadPassword(prompt)
	panicIfFailed(err)
	return res
}
