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

// Package gcp contains helpers backed by Google Cloud services. This file
// implements datastore-backed storage of authenticated users and their roles.
package gcp

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/v2/datastore"
)

// User represents an authenticated user.
// User represents an authenticated user stored in the datastore.
type User struct {
	Email   string    // unique user identifier and datastore key
	Role    string    // current user role, e.g. "admin" or "organizer"
	Created time.Time // timestamp when the entity was first stored
}

// EnsureUserExists retrieves a user from the datastore by email. If the user
// does not exist a new entity is created with the provided role. If the user
// exists but the role differs, the stored role is updated. The resulting user
// entity is returned.
func EnsureUserExists(c context.Context, email, role string) (*User, error) {
	key := datastore.NewKey(c, "User", email, 0, nil)
	var u User
	err := datastore.Get(c, key, &u)
	if err == datastore.ErrNoSuchEntity {
		u = User{Email: email, Role: role, Created: time.Now()}
		if _, err := datastore.Put(c, key, &u); err != nil {
			return nil, err
		}
		return &u, nil
	}
	if err != nil {
		return nil, err
	}
	if u.Role != role {
		u.Role = role
		if _, err := datastore.Put(c, key, &u); err != nil {
			return nil, err
		}
	}
	return &u, nil
}

// GetUserRole fetches and returns the role for the user with the given email.
// It returns datastore.ErrNoSuchEntity if the user is not found.
func GetUserRole(c context.Context, email string) (string, error) {
	key := datastore.NewKey(c, "User", email, 0, nil)
	var u User
	if err := datastore.Get(c, key, &u); err != nil {
		return "", err
	}
	return u.Role, nil
}
