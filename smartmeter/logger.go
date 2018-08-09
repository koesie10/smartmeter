package smartmeter

import (
	"bytes"
	"fmt"
	"io"
	"os"
		"time"
)

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
}

func NewStdoutLog() Logger {
	return &writerLogger{
		w: os.Stdout,
	}
}

func NewStderrLog() Logger {
	return &writerLogger{
		w: os.Stderr,
	}
}

type writerLogger struct {
	w io.Writer
}

func (l *writerLogger) Debug(format string, args ...interface{}) {
	l.log("DEBUG", format, args...)
}

func (l *writerLogger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

func (l *writerLogger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

func (l *writerLogger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

func (l *writerLogger) log(level string, format string, args ...interface{}) {
	s := bytes.NewBufferString("[")
	s.WriteString(level)
	s.WriteRune(']')
	s.WriteRune(' ')
	s.WriteString(time.Now().Format(time.RFC3339Nano))
	s.WriteRune(' ')
	s.WriteString(fmt.Sprintf(format, args...))
	s.WriteString("\n")
	l.w.Write(s.Bytes())
}
