package main

// Cache has two interchangeable implementations, told apart by name.
type Cache interface {
	Name() string
}

type redisCache struct{}

func NewRedisCache() *redisCache { return &redisCache{} }
func (*redisCache) Name() string { return "redis" }

type memoryCache struct{}

func NewMemoryCache() *memoryCache { return &memoryCache{} }
func (*memoryCache) Name() string  { return "memory" }

func main() {}
