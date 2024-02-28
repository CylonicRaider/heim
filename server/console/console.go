package console

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"

	"github.com/euphoria-io/scope"
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

type ConsoleWithContext interface {
	Console
	Context() scope.Context
}

type readRequest struct {
	prompt string
	hidden bool
	result chan<- readResponse
}

type readResponse struct {
	text string
	err  error
}

type DefaultConsoleWithContext struct {
	*DefaultConsole
	ctx      scope.Context
	requests chan readRequest
}

func (c *DefaultConsole) WithContext(ctx scope.Context) *DefaultConsoleWithContext {
	result := &DefaultConsoleWithContext{
		DefaultConsole: c,
		ctx:            ctx,
		requests:       make(chan readRequest, 16),
	}
	go result.background()
	return result
}

func (c *DefaultConsoleWithContext) Context() scope.Context {
	return c.ctx
}

func (c *DefaultConsoleWithContext) background() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case req := <-c.requests:
			var text string
			var err error
			if req.hidden {
				text, err = c.DefaultConsole.ReadPassword(req.prompt)
			} else {
				text, err = c.DefaultConsole.ReadLine(req.prompt)
			}
			req.result <- readResponse{text, err}
		}
	}
}

func (c *DefaultConsoleWithContext) read(prompt string, hidden bool) (string, error) {
	if !c.ctx.Alive() {
		return "", c.ctx.Err()
	}
	back := make(chan readResponse, 1)
	req := readRequest{prompt, hidden, back}
	select {
	case <-c.ctx.Done():
		return "", c.ctx.Err()
	case c.requests <- req:
	}
	select {
	case <-c.ctx.Done():
		// Try to suppress at least a little visual noise from a still-ongoing read.
		// Any unconsumed input may still get reprinted.
		c.term.SetPrompt("")
		return "", c.ctx.Err()
	case resp := <-back:
		return resp.text, resp.err
	}
}

func (c *DefaultConsoleWithContext) ReadLine(prompt string) (string, error) {
	return c.read(prompt, false)
}

func (c *DefaultConsoleWithContext) ReadPassword(prompt string) (string, error) {
	return c.read(prompt, true)
}
