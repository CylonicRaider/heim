package logging

import (
	"io"
	"log"
	"os"

	"github.com/euphoria-io/scope"
)

type logCtxKey int

const ctxLogger logCtxKey = 0
const ctxWriter logCtxKey = 1

const logFlags = log.LstdFlags

func Logger(ctx scope.Context) *log.Logger {
	if logger, ok := ctx.Get(ctxLogger).(*log.Logger); ok {
		return logger
	}
	return log.New(GetDefaultWriterOrFallback(ctx, os.Stdout), "[???] ", logFlags)
}

func GetDefaultWriter(ctx scope.Context) io.Writer {
	w, _ := ctx.Get(ctxWriter).(io.Writer)
	return w
}

func GetDefaultWriterOrFallback(ctx scope.Context, w io.Writer) io.Writer {
	if cw, ok := ctx.Get(ctxWriter).(io.Writer); ok {
		w = cw
	}
	return w
}

func SetDefaultWriter(ctx scope.Context, w io.Writer) scope.Context {
	ctx.Set(ctxWriter, w)
	return ctx
}

func LoggingContextOverride(ctx scope.Context, w io.Writer, prefix string) scope.Context {
	logger := log.New(w, prefix, logFlags)
	ctx.Set(ctxLogger, logger)
	ctx.Set(ctxWriter, w)
	return ctx
}

func LoggingContext(ctx scope.Context, w io.Writer, prefix string) scope.Context {
	return LoggingContextOverride(ctx, GetDefaultWriterOrFallback(ctx, w), prefix)
}
