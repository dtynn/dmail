package safeMap

import (
	"testing"
	"time"
)

func TestSafeMap(t *testing.T) {
	s := NewSafeMap()

	key := "key"
	keyExpire := "key_expire"
	keyNotFound := "key_not_found"
	val := "val"
	// set
	err := s.Set(key, val)
	if err != nil {
		t.Error("Set: ", err)
	}

	// get
	v, err := s.Get(key)
	if err != nil {
		t.Error("Get: ", err)
	}
	if v != val {
		t.Errorf("Expect %s , get %s", val, v)
	}

	// get not found
	if _, err = s.Get(keyNotFound); err != errCacheNotFound {
		t.Errorf("Expect not found, get %s", err)
	}

	if err = s.Setex(keyExpire, val, 3); err != nil {
		t.Error("Setex: ", err)
	}

	if v, _ = s.Get(keyExpire); v != val {
		t.Errorf("Get Expire Cache: expect %s , get %s", val, v)
	}

	time.Sleep(time.Duration(5) * time.Second)
	if _, err = s.Get(keyExpire); err != errCacheNotFound {
		t.Errorf("Expect error %s , get %s", errCacheNotFound, err)
	}

	if err = s.Set(key, val); err != nil {
		t.Error("Set: ", err)
	}

	if err = s.Setnx(key, val); err != errCacheExists {
		t.Errorf("Expect error exists, get %s", err)
	}

}
