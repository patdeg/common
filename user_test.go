package common

import (
	"testing"

	"google.golang.org/appengine/v2/aetest"
)

func TestEnsureUserExistsUpdatesRole(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Skipf("aetest not available: %v", err)
		return
	}
	defer done()

	email := "test@example.com"
	if _, err := EnsureUserExists(ctx, email, "organizer"); err != nil {
		t.Fatalf("EnsureUserExists: %v", err)
	}
	if role, _ := GetUserRole(ctx, email); role != "organizer" {
		t.Fatalf("got role %q, want organizer", role)
	}
	if _, err := EnsureUserExists(ctx, email, "admin"); err != nil {
		t.Fatalf("EnsureUserExists update: %v", err)
	}
	role, err := GetUserRole(ctx, email)
	if err != nil {
		t.Fatalf("GetUserRole: %v", err)
	}
	if role != "admin" {
		t.Errorf("role not updated, got %q", role)
	}
}
