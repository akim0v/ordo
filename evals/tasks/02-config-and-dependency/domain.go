package main

// Config is a settings value the application constructs itself.
type Config struct {
	Prefix string
}

// Store is the storage boundary.
type Store interface {
	Kind() string
}

type memStore struct{}

// NewMemStore constructs the Store implementation.
func NewMemStore() *memStore { return &memStore{} }

func (*memStore) Kind() string { return "mem" }

// Service needs configuration and a store.
type Service struct {
	cfg   *Config
	store Store
}

// NewService takes two dependencies.
func NewService(cfg *Config, store Store) *Service {
	return &Service{cfg: cfg, store: store}
}

func (s *Service) Describe() string { return s.cfg.Prefix + ":" + s.store.Kind() }

func main() {}
