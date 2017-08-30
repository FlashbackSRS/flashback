// +build !debug

package log

// Debug calls log.Print() only when the 'debug' build flag is present.
func Debug(v ...interface{}) {}

// Debugf calls log.Printf() only when the 'debug' build flag is present.
func Debugf(format string, v ...interface{}) {}

// Debugln calls log.Println() only when the 'debug' build flag is present.
func Debugln(v ...interface{}) {}

// Debug calls log.Print() only when the 'debug' build flag is present.
func (l *Logger) Debug(v ...interface{}) {}

// Debugf calls log.Printf() only when the 'debug' build flag is present.
func (l *Logger) Debugf(format string, v ...interface{}) {}

// Debugln calls log.Println() only when the 'debug' build flag is present.
func (l *Logger) Debugln(v ...interface{}) {}
