package main

// Store is the boundary Service depends on.
type Store interface {
	Kind() string
}

type memStore struct{}

// NewMemStore constructs the Store implementation.
func NewMemStore() *memStore { return &memStore{} }

func (*memStore) Kind() string { return "mem" }

// Service needs a Store.
type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Kind() string { return s.store.Kind() }

func main() {}
