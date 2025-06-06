package common

import "log"

// Debug logs a formatted message when ISDEBUG is true.
func Debug(format string, v ...interface{}) {
	if !ISDEBUG {
		return
	}
	log.Printf(format+"\n", v...)
}

// Info logs a formatted informational message.
func Info(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

// Error logs a formatted error message prefixed with "ERROR:".
func Error(format string, v ...interface{}) {
	log.Printf("ERROR: "+format+"\n", v...)
}
