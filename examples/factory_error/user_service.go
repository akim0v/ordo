package main

// UserRepository is the dependency UserService is built from.
type UserRepository interface {
	GetUserName(id int) (string, bool)
}

// UserService depends on a UserRepository it never gets, because the
// repository constructor fails.
type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}
