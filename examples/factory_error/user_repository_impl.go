package main

import "errors"

// ErrNoDatabase stands in for the connection failure a real repository
// constructor would report.
var ErrNoDatabase = errors.New("database is unreachable")

type UserRepositoryImpl struct{}

var _ UserRepository = (*UserRepositoryImpl)(nil)

// NewUserRepositoryImpl returns (T, error), the second supported factory shape.
// Returning a non-nil error fails the resolution that needed this service.
func NewUserRepositoryImpl() (*UserRepositoryImpl, error) {
	return nil, ErrNoDatabase
}

func (ur *UserRepositoryImpl) GetUserName(int) (string, bool) {
	return "", false
}
