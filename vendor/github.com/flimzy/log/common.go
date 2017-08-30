package log

import (
	"io"
	golog "log"
)

// A Logger represents an active logging object that generates lines of output
// to an io.Writer. Each logging operation makes a single call to the Writer's
// Write method. A Logger can be used simultaneously from multiple goroutines;
// it guarantees to serialize access to the Writer.
type Logger struct {
	golog *golog.Logger
}

// New creates a new Logger.
func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{golog.New(out, prefix, flag)}
}

// Flags returns the output flags for the standard logger.
func Flags() int {
	return golog.Flags()
}

// Print calls log.Print()
func Print(v ...interface{}) {
	golog.Print(v...)
}

// Printf calls log.Printf()
func Printf(format string, v ...interface{}) {
	golog.Printf(format, v...)
}

// Println calls log.Println()
func Println(v ...interface{}) {
	golog.Println(v...)
}

// Print calls log.Print()
func (l *Logger) Print(v ...interface{}) {
	l.golog.Print(v...)
}

// Printf calls log.Printf()
func (l *Logger) Printf(format string, v ...interface{}) {
	l.golog.Printf(format, v...)
}

// Println calls log.Println()
func (l *Logger) Println(v ...interface{}) {
	l.golog.Println(v...)
}

// Flags returns the output flags for the logger.
func (l *Logger) Flags() int {
	return l.golog.Flags()
}
