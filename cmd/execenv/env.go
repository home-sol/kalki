// Package execenv provides the environment for executing commands.
package execenv

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// RootCommandName is the name of the root command.
const RootCommandName = "kalki"

// Env is the environment of a command.
type Env struct {
	Logger *zap.Logger
	Out    Out
	Err    Out
}

// NewEnv creates a new environment.
func NewEnv() *Env {
	zapOptions := []zap.Option{
		zap.AddStacktrace(zapcore.FatalLevel),
		zap.AddCallerSkip(3),
	}

	zapOptions = append(zapOptions,
		zap.IncreaseLevel(zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l != zapcore.DebugLevel })),
	)
	logger, _ := zap.NewDevelopment(zapOptions...)
	return &Env{
		Logger: logger,
		Out:    out{Writer: os.Stdout},
		Err:    out{Writer: os.Stderr},
	}
}

// Out is the output interface.
type Out interface {
	io.Writer
	Printf(format string, a ...interface{})
	Print(a ...interface{})
	Println(a ...interface{})
}

type out struct {
	io.Writer
}

func (o out) Printf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(o, format, a...)
}

func (o out) Print(a ...interface{}) {
	_, _ = fmt.Fprint(o, a...)
}

func (o out) Println(a ...interface{}) {
	_, _ = fmt.Fprintln(o, a...)
}
