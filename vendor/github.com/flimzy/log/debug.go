// +build debug

package log

import (
	golog "log"
)

// Debug calls log.Print() only when the 'debug' build flag is present.
func Debug(v ...interface{}) {
	golog.Print(v...)
}

// Debugf calls log.Printf() only when the 'debug' build flag is present.
func Debugf(format string, v ...interface{}) {
	golog.Printf(format, v...)
}

// Debugln calls log.Println() only when the 'debug' build flag is present.
func Debugln(v ...interface{}) {
	golog.Println(v...)
}

// Debug calls log.Print() only when the 'debug' build flag is present.
func (l *Logger) Debug(v ...interface{}) {
	l.Print(v...)
}

// Debugf calls log.Printf() only when the 'debug' build flag is present.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Printf(format, v...)
}

// Debugln calls log.Println() only when the 'debug' build flag is present.
func (l *Logger) Debugln(v ...interface{}) {
	l.Println(v...)
}
