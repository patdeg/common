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

package gcp

import (
	"strings"

	"golang.org/x/net/context"
	appengine "google.golang.org/appengine/v2"
)

// Version retrieves the App Engine version ID from context and stores the
// major component in the exported `common.VERSION` variable. This helper should
// be called once during initialization so callers can read `common.VERSION`
// throughout the application.
func Version(c context.Context) string {
	version := appengine.VersionID(c)
	array := strings.Split(version, ".")
	VERSION = array[0]
	return VERSION
}
