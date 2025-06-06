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
