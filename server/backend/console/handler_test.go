package console

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/euphoria-io/scope"

	. "github.com/smartystreets/goconvey/convey"
)

type testTerm struct {
	bytes.Buffer
	password string
}

func (t *testTerm) ReadPassword(password string) (string, error) { return t.password, nil }

type testHandler struct{}

func (testHandler) run(ctx scope.Context, c *console, args []string) error {
	if len(args) != 1 {
		return usageError("invalid number of arguments: %d", len(args))
	}
	fmt.Fprintln(c, "ok")
	return nil
}

type testHandlerWithUsage struct {
	testHandler
}

func (testHandlerWithUsage) usage() string { return "usage" }

func TestRunHandler(t *testing.T) {
	ctrl := &Controller{}
	ctx := scope.New()

	Convey("Successfully runs", t, func() {
		term := &testTerm{}
		runHandler(ctx, testHandler{}, cmdConsole(ctrl, "test", term), []string{"test"})
		So(term.String(), ShouldEqual, "ok\r\n")
	})

	Convey("Usage error", t, func() {
		Convey("Handler serves usage", func() {
			term := &testTerm{}
			runHandler(ctx, testHandlerWithUsage{}, cmdConsole(ctrl, "test", term), nil)
			So(term.String(), ShouldEqual,
				"error: invalid number of arguments: 0\r\nusage\r\n\r\nOPTIONS:\r\n")
		})

		Convey("Handler doesn't serve usage", func() {
			term := &testTerm{}
			runHandler(ctx, testHandler{}, cmdConsole(ctrl, "test", term), nil)
			So(term.String(), ShouldEqual, "error: invalid number of arguments: 0\r\n")
		})
	})
}

func TestRunCommand(t *testing.T) {
	ctx := scope.New()

	Convey("Unregistered command prints error", t, func() {
		term := &testTerm{}
		runCommand(ctx, nil, "asdf", term, nil)
		So(term.String(), ShouldEqual, "invalid command: asdf\r\n")
	})

	Convey("Registered command is invoked", t, func() {
		save := handlers
		defer func() { handlers = save }()
		handlers = map[string]handler{}
		register("test", testHandler{})
		term := &testTerm{}
		runCommand(ctx, &Controller{}, "test", term, []string{"arg"})
		So(term.String(), ShouldEqual, "ok\r\n")
	})
}
