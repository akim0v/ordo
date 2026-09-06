package main

// UserService is the usecase layer. It depends on the UserRepository interface
// and knows nothing about the implementation behind it.
type UserService struct {
	repository UserRepository
	nextID     int
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository, nextID: 1}
}

func (s *UserService) Create(name string) User {
	user := User{ID: s.nextID, Name: name}
	s.nextID++

	s.repository.Insert(user)
	return user
}

func (s *UserService) Get(id int) (User, bool) {
	return s.repository.Get(id)
}
