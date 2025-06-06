package common

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/v2/datastore"
)

func GetFirst(c context.Context, q *datastore.Query, dst interface{}) (*datastore.Key, error) {
	t := q.Run(c)
	key, err := t.Next(dst)
	if err == datastore.Done {
		// no results
		return nil, datastore.ErrNoSuchEntity
	}
	if err != nil {
		// error
		return nil, err
	}
	return key, nil
}
