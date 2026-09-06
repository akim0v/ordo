package main

// Store is the boundary a test substitutes.
type Store interface {
	Kind() string
}

type realStore struct{}

// NewRealStore constructs the production implementation.
func NewRealStore() *realStore { return &realStore{} }

func (*realStore) Kind() string { return "real" }

// Service depends on the interface, never the implementation.
type Service struct {
	store Store
}

func NewService(store Store) *Service { return &Service{store: store} }

func (s *Service) Kind() string { return s.store.Kind() }

func main() {}
