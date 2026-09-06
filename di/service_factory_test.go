package di

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// NewServiceFactorySuite is a test suite for the newServiceFactory function
type NewServiceFactorySuite struct {
	suite.Suite
}

// TestUsual tests the usual case
func (suite *NewServiceFactorySuite) TestUsual() {
	// Arrange
	type MyStruct struct{}

	// Act
	f, err := newServiceFactory(func(int, string) (s MyStruct) {
		return
	})

	// Assert
	suite.Require().NoError(err)
	if suite.NotNil(f) {
		suite.Equal(2, f.DepsCount)
	}
}

// TestWithError tests the constructor with an error returned
func (suite *NewServiceFactorySuite) TestWithError() {
	// Arrange
	type MyStruct struct{}

	// Act
	f, err := newServiceFactory(func() (s MyStruct, err error) {
		return
	})

	// Assert
	suite.Require().NoError(err)
	if suite.NotNil(f) {
		suite.True(f.HasErr)
	}
}

// TestNonErrorSecondReturnArgument tests the constructor with a non-error
// second return argument
func (suite *NewServiceFactorySuite) TestNonErrorSecondReturnArgument() {
	// Arrange
	type (
		MyStruct  struct{}
		MyStruct2 struct{}
	)

	// Act
	f, err := newServiceFactory(func() (s MyStruct, s2 MyStruct2) {
		return
	})

	// Assert
	suite.Nil(f)
	suite.ErrorIs(err, ErrFactorySecondReturnNotErr)
}

// TestNonFunctionArgument tests the case when client provides a non-function argument
func (suite *NewServiceFactorySuite) TestNonFunctionArgument() {
	// Arrange
	args := []any{0.1, 1, "string", struct{}{}}

	for _, arg := range args {
		// Act
		f, err := newServiceFactory(arg)

		// Assert
		suite.Nil(f)
		suite.ErrorIs(err, ErrFactoryNotFunction)
	}
}

// TestNilArgument tests the case when client provides a nil argument
func (suite *NewServiceFactorySuite) TestNilArgument() {
	// Act
	f, err := newServiceFactory(nil)

	// Assert
	suite.Nil(f)
	suite.ErrorIs(err, ErrNilRegistration)
}

// TestNoReturnArguments tests the case the provided function has no return arguments
func (suite *NewServiceFactorySuite) TestNoReturnArguments() {
	// Act
	f, err := newServiceFactory(func() {})

	// Assert
	suite.Nil(f)
	suite.ErrorIs(err, ErrFactoryNoReturn)
}

// TestTooMuchReturnArguments tests the case when the function has too many return arguments
func (suite *NewServiceFactorySuite) TestTooMuchReturnArguments() {
	// Act
	f, err := newServiceFactory(func() (v1 int, err error, v2 int) {
		return
	})

	// Assert
	suite.Nil(f)
	suite.ErrorIs(err, ErrFactoryTooManyReturns)
}

// TestNewServiceFactory tests the newServiceFactory function
func TestNewServiceFactory(t *testing.T) {
	suite.Run(t, new(NewServiceFactorySuite))
}
