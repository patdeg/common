// Package gcp provides Google Cloud helpers. This file defines helpers for
// retrieving results from Cloud Datastore queries.
package gcp

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/v2/datastore"
)

// GetFirst executes the query and loads the first entity into dst.
//
// Parameters:
//
//	c - context used for datastore calls.
//	q - query to run.
//	dst - pointer that will receive the entity data.
//
// Returns the entity key if one was found. When the query yields no results,
// the key is nil and the error is datastore.ErrNoSuchEntity. Any other error
// encountered when fetching the first result is returned as-is.
// The datastore.Done error from t.Next is converted to datastore.ErrNoSuchEntity
// so callers can easily detect the no-results case.
func GetFirst(c context.Context, q *datastore.Query, dst interface{}) (*datastore.Key, error) {
	t := q.Run(c)
	key, err := t.Next(dst)
	if err == datastore.Done {
		return nil, datastore.ErrNoSuchEntity
	}
	if err != nil {
		return nil, err
	}
	return key, nil
}
