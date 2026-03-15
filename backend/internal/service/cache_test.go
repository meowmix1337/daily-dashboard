package service

import (
	"testing"
	"time"
)

func TestCacheService_SetAndGet(t *testing.T) {
	c := &CacheService{}

	c.Set("key1", "value1", time.Minute)
	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v.(string) != "value1" {
		t.Fatalf("got %v, want value1", v)
	}
}

func TestCacheService_Miss(t *testing.T) {
	c := &CacheService{}

	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCacheService_Expiry(t *testing.T) {
	c := &CacheService{}

	c.Set("key", "val", -time.Second) // already expired
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected expired entry to be a miss")
	}
}

func TestCacheService_Delete(t *testing.T) {
	c := &CacheService{}

	c.Set("key", "val", time.Minute)
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}

func TestCacheService_Overwrite(t *testing.T) {
	c := &CacheService{}

	c.Set("key", "first", time.Minute)
	c.Set("key", "second", time.Minute)
	v, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v.(string) != "second" {
		t.Fatalf("got %v, want second", v)
	}
}
