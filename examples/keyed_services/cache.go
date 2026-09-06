package main

import "fmt"

// Cache is one interface with several interchangeable implementations, told
// apart by name rather than by type.
type Cache interface {
	Name() string
}

type RedisCache struct{}

func NewRedisCache() *RedisCache { return &RedisCache{} }
func (*RedisCache) Name() string { return "redis" }

type MemoryCache struct{}

func NewMemoryCache() *MemoryCache { return &MemoryCache{} }
func (*MemoryCache) Name() string  { return "memory" }

// NullCache is registered as an already-built value.
type NullCache struct{}

func (*NullCache) Name() string { return "null" }

// SessionStore takes a plain Cache parameter. Its dependency is resolved by
// type, with no key involved.
type SessionStore struct {
	cache Cache
}

func NewSessionStore(cache Cache) *SessionStore {
	return &SessionStore{cache: cache}
}

func (s *SessionStore) Describe() string {
	return fmt.Sprintf("session store backed by %s", s.cache.Name())
}
