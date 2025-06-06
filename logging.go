// logging.go provides a tiny wrapper around the standard log package.
// The helpers here format messages consistently and funnel all logs
// through log.Printf. Debug obeys the global ISDEBUG variable so it
// can be disabled globally.

package common

import "log"

// Debug writes a formatted debug message when ISDEBUG is true.
// A newline is appended so callers do not have to include one.
func Debug(format string, v ...interface{}) {
	if !ISDEBUG {
		return
	}
	// Include trailing newline for consistency with other helpers.
	log.Printf(format+"\n", v...)
}

// Info writes a formatted informational message.
// A newline is appended to keep log lines consistent.
func Info(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

// Error writes a formatted error message with an "ERROR:" prefix.
// The prefix helps grep for errors in log files.
func Error(format string, v ...interface{}) {
	log.Printf("ERROR: "+format+"\n", v...)
}
