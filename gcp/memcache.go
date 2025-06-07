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

// This file provides small wrappers around App Engine's memcache library.
// Cached values are stored in plain text and may be evicted at any time, so
// do not store sensitive information unless it is encrypted.

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/v2/memcache"
)

// DeleteMemCache removes a key from memcache.
// memcache.ErrCacheMiss is ignored so callers can treat missing keys as a no-op.
// Any other error is logged and returned.
func DeleteMemCache(c context.Context, key string) (err error) {
	err = memcache.Delete(c, key)
	if err == memcache.ErrCacheMiss {
		return nil
	}
	if err != nil {
		Error("GetMemCache: error getting item %v: %v", key, err)
		return err
	}
	return nil
}

// GetMemCache retrieves raw bytes from memcache.
// When the key is missing, memcache.ErrCacheMiss is returned with an empty slice.
// Other errors are logged and returned with an empty slice.
func GetMemCache(c context.Context, key string) ([]byte, error) {
	object, err := memcache.Get(c, key)
	if err == memcache.ErrCacheMiss {
		return []byte{}, err
	}
	if err != nil {
		Error("GetMemCache: error getting item %v: %v", key, err)
		return []byte{}, err
	}
	return object.Value, nil
}

// SetMemCache stores bytes in memcache for the specified number of hours.
// If the key already exists it is overwritten. Errors are logged.
func SetMemCache(c context.Context, key string, item []byte, hours int32) {
	object := &memcache.Item{
		Key:        key,
		Value:      item,
		Expiration: time.Hour * time.Duration(hours),
	}
	// Add the item to the memcache, if the key does not already exist
	if err := memcache.Add(c, object); err == memcache.ErrNotStored {
		Info("SetMemCache: item %q already exists", key)
		// Update content
		if err := memcache.Set(c, object); err != nil {
			Error("SetMemCache: Error updating memcache item %q: %v", key, err)
		}
	} else if err != nil {
		Error("SetMemCache: error adding item %q: %v", key, err)
	}
}

// GetMemCacheString returns the cached string for the provided key.
// An empty string is returned when the key is missing or an error occurs.
//
// Example:
//
//	name := GetMemCacheString(ctx, "user-name")
func GetMemCacheString(c context.Context, key string) string {
	item, err := GetMemCache(c, key)
	if err != nil {
		return ""
	}
	return B2S(item)
}

// SetMemCacheString stores the provided string in memcache for the given hours.
// Existing values are overwritten.
//
// Example:
//
//	SetMemCacheString(ctx, "user-name", "gopher", 2)
func SetMemCacheString(c context.Context, key string, item string, hours int32) {
	SetMemCache(c, key, []byte(item), hours)
}

// GetObjMemCache retrieves an object stored with SetObjMemCache.
// memcache.ErrCacheMiss is returned when the key is missing. Any other error is
// returned as-is.
func GetObjMemCache(c context.Context, key string, v interface{}) error {
	_, err := memcache.Gob.Get(c, key, v)
	return err
}

// SetObjMemCache stores an arbitrary object in memcache for the given duration.
// Do not store sensitive data here unless it is encrypted as the cache is
// accessible by other applications running in the same environment.
func SetObjMemCache(c context.Context, key string, v interface{}, hours int32) error {
	item := memcache.Item{
		Key:        key,
		Object:     v,
		Expiration: time.Hour * time.Duration(hours),
	}
	return memcache.Gob.Set(c, &item)
}
