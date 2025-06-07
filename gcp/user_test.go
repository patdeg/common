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
	"context"
	"testing"

	"google.golang.org/appengine/v2/aetest"
	"google.golang.org/appengine/v2/datastore"
)

// TestEnsureUserExistsUpdatesRole verifies that calling EnsureUserExists on an
// existing user updates the stored role.
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

// TestEnsureUserExistsCreatesUser ensures a new user entity is created with the
// provided role when none exists.
func TestEnsureUserExistsCreatesUser(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Skipf("aetest not available: %v", err)
		return
	}
	defer done()

	email := "new@example.com"
	if _, err := EnsureUserExists(ctx, email, "member"); err != nil {
		t.Fatalf("EnsureUserExists: %v", err)
	}
	role, err := GetUserRole(ctx, email)
	if err != nil {
		t.Fatalf("GetUserRole: %v", err)
	}
	if role != "member" {
		t.Errorf("got role %q, want member", role)
	}
}

// TestGetUserRoleNotFound verifies that requesting the role of a nonexistent
// user returns datastore.ErrNoSuchEntity.
func TestGetUserRoleNotFound(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Skipf("aetest not available: %v", err)
		return
	}
	defer done()

	_, err = GetUserRole(ctx, "missing@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != datastore.ErrNoSuchEntity {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestEnsureUserExistsBadContext confirms an error is returned when the
// datastore service is unavailable.
func TestEnsureUserExistsBadContext(t *testing.T) {
	// Use a background context without the aetest datastore service.
	if _, err := EnsureUserExists(context.Background(), "bad@example.com", "admin"); err == nil {
		t.Fatal("expected error with background context")
	}
}
