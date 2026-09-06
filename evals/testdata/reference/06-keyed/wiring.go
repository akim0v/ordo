package main

import "github.com/akim0v/ordo"

func BuildContainer() (*ordo.Container, error) {
	return ordo.New(
		ordo.WithKeyedService[Cache]("redis", NewRedisCache),
		ordo.WithKeyedService[Cache]("memory", NewMemoryCache),
	)
}

func Get(c *ordo.Container, key string) (Cache, error) {
	return c.GetKeyedService[Cache](key)
}
