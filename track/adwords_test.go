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

package track

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestAdWordsTrackingHandlerInvalidURL verifies invalid redirect URLs return 400.
func TestAdWordsTrackingHandlerInvalidURL(t *testing.T) {
	os.Setenv("GAE_INSTANCE", "test")
	os.Setenv("GAE_VERSION", "1")
	os.Setenv("GAE_DEPLOYMENT_ID", "1")
	r := httptest.NewRequest("GET", "/tracking?url=invalid", nil)
	w := httptest.NewRecorder()
	AdWordsTrackingHandler(w, r)
	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusBadRequest)
	}
}
