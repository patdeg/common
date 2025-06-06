// Package common provides helper utilities shared across the repository. The
// helpers include logging routines, conversion helpers and small web utilities.
//
// Example:
//
//	if common.ISDEBUG {
//	    common.Debug("processing request")
//	}
package common

import (
	"os"
	"strconv"
)

var (
	// ISDEBUG reports whether debug logs should be printed. The value is
	// derived from the COMMON_DEBUG environment variable. When unset or not
	// a valid boolean, it defaults to false.
	//
	// Example:
	//
	//     $ COMMON_DEBUG=true ./your_binary
	//     if common.ISDEBUG {
	//         common.Debug("debugging enabled")
	//     }
	ISDEBUG = func() bool {
		v, err := strconv.ParseBool(os.Getenv("COMMON_DEBUG"))
		return err == nil && v
	}()

	// VERSION stores the application version. It can be overridden at build
	// time using -ldflags:
	//
	//     go build -ldflags "-X github.com/patdeg/common.VERSION=1.0.0"
	VERSION string
)
