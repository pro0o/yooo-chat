package slog

import (
	"fmt"
	"io"
)

type Logger struct {
	out io.Writer
}

func New(out io.Writer) *Logger {
	return &Logger{out: out}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(l.out, format, v...)
}
