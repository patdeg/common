package common

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/v2/memcache"
)

// Wrapper around Delete Memcache.
func DeleteMemCache(c context.Context, key string) (err error) {
	err = memcache.Delete(c, key)
	if err == memcache.ErrCacheMiss {
		err = nil
		return
	} else if err != nil {
		Error("GetMemCache: error getting item %v: %v", key, err)
		return
	}
	return
}

// Wrapper around Get Memcache. Return empty if error or not found.
func GetMemCache(c context.Context, key string) ([]byte, error) {
	object, err := memcache.Get(c, key)
	if err == memcache.ErrCacheMiss {
		return []byte{}, err
	} else if err != nil {
		Error("GetMemCache: error getting item %v: %v", key, err)
		return []byte{}, err
	}
	return object.Value, nil
}

// Wrapper around Set Memcache. Return empty if error or not found.
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

func GetMemCacheString(c context.Context, key string) string {
	item, err := GetMemCache(c, key)
	if err != nil {
		return ""
	}
	return B2S(item)
}

func SetMemCacheString(c context.Context, key string, item string, hours int32) {
	SetMemCache(c, key, []byte(item), hours)
}

// Wrapper around Get Memcache. Return memcache.ErrCacheMiss if not found, or error
func GetObjMemCache(c context.Context, key string, v interface{}) error {
	_, err := memcache.Gob.Get(c, key, v)
	return err
}

// Wrapper around Set Memcache Object
func SetObjMemCache(c context.Context, key string, v interface{}, hours int32) error {
	item := memcache.Item{
		Key:        key,
		Object:     v,
		Expiration: time.Hour * time.Duration(hours),
	}
	return memcache.Gob.Set(c, &item)
}
