package main

import "log"

type User struct {
	ID   int
	Name string
}

// UserRepository is the boundary the usecase layer depends on. The storage
// implementation below is the only thing that knows it is a map.
type UserRepository interface {
	Insert(User)
	Get(id int) (User, bool)
}

type InMemoryUserRepository struct {
	debug bool
	users map[int]User
}

var _ UserRepository = (*InMemoryUserRepository)(nil)

// NewInMemoryUserRepository depends on *Config, which the container supplies
// from the value registered in main.
func NewInMemoryUserRepository(cfg *Config) *InMemoryUserRepository {
	return &InMemoryUserRepository{
		debug: cfg.Debug,
		users: make(map[int]User),
	}
}

func (r *InMemoryUserRepository) Insert(user User) {
	if r.debug {
		log.Printf("storage: insert user %d (%s)", user.ID, user.Name)
	}
	r.users[user.ID] = user
}

func (r *InMemoryUserRepository) Get(id int) (User, bool) {
	user, ok := r.users[id]
	return user, ok
}
