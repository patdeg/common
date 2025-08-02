// Copyright 2025 Patrick Deglon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// Warn writes a formatted warning message with an "WARNING:" prefix.
// The prefix helps grep for warnings in log files.
func Warn(format string, v ...interface{}) {
	log.Printf("WARNING: "+format+"\n", v...)
}

// Error writes a formatted error message with an "ERROR:" prefix.
// The prefix helps grep for errors in log files.
func Error(format string, v ...interface{}) {
	log.Printf("ERROR: "+format+"\n", v...)
}
