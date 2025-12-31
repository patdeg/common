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

	// LLMAPIKey stores the API key used to contact the default LLM provider.
	// Users should set the COMMON_LLM_API_KEY environment variable prior to
	// running their binaries.
	LLMAPIKey = os.Getenv("COMMON_LLM_API_KEY")

	// LLMModel defines the default LLM model identifier. The value may be
	// overridden via the COMMON_LLM_MODEL environment variable.
	LLMModel = func() string {
		if v := os.Getenv("COMMON_LLM_MODEL"); v != "" {
			return v
		}
		return "meta-llama/llama-4-scout-17b-16e-instruct"
	}()

	// LLMBaseURL defines the HTTP endpoint for the LLM provider. It defaults to
	// Demeterics' Groq proxy but can be overridden with COMMON_LLM_BASE_URL.
	LLMBaseURL = func() string {
		if v := os.Getenv("COMMON_LLM_BASE_URL"); v != "" {
			return v
		}
		return "https://api.demeterics.com/groq/v1"
	}()
)
