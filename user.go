package common

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/v2/datastore"
)

// User represents an authenticated user.
type User struct {
	Email   string
	Role    string
	Created time.Time
}

// EnsureUserExists checks if a User entity with the given email exists. If not,
// it is created with the provided role.
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

// GetUserRole fetches the role for the given user email.
func GetUserRole(c context.Context, email string) (string, error) {
	key := datastore.NewKey(c, "User", email, 0, nil)
	var u User
	if err := datastore.Get(c, key, &u); err != nil {
		return "", err
	}
	return u.Role, nil
}
