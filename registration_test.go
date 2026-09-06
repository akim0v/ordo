package ordo

import (
	"errors"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
)

// registrationTestFile is the base name of this file, reported by every call
// site captured in it
const registrationTestFile = "registration_test.go"

// badRegistrationFromHelper returns a malformed registration created away from
// the New call, together with the line the registration was created on
func badRegistrationFromHelper() (Option, int) {
	opt := WithService[testRepository]("not a factory")
	_, _, line, _ := runtime.Caller(0)

	// The registration is created on the line right above the call reporting it
	return opt, line - 1
}

// CallSiteSuite is the suite for testing the callSite type
type CallSiteSuite struct {
	suite.Suite
}

// TestCapturesCaller tests the helper reports the source location of its caller
func (suite *CallSiteSuite) TestCapturesCaller() {
	// Act
	site := newCallSite(0)
	_, _, line, _ := runtime.Caller(0)

	// Assert
	suite.False(site.IsZero())
	suite.Equal(registrationTestFile, filepath.Base(site.File))
	suite.NotZero(site.Line)
	suite.Equal(line-1, site.Line)
	suite.Contains(site.String(), registrationTestFile)
}

// TestSkipsFrames tests the helper ascends the requested number of frames
func (suite *CallSiteSuite) TestSkipsFrames() {
	// Arrange
	nested := func() callSite {
		return newCallSite(1)
	}

	// Act
	site := nested()
	_, _, line, _ := runtime.Caller(0)

	// Assert
	suite.Equal(registrationTestFile, filepath.Base(site.File))
	suite.Equal(line-1, site.Line)
}

// TestZeroSiteRendersUnknown tests an uncaptured call site renders a placeholder
func (suite *CallSiteSuite) TestZeroSiteRendersUnknown() {
	// Arrange
	var site callSite

	// Act & Assert
	suite.True(site.IsZero())
	suite.Equal("unknown source", site.String())
}

// TestCallSite tests the callSite type
func TestCallSite(t *testing.T) {
	suite.Run(t, new(CallSiteSuite))
}

// InvalidRegistrationErrorSuite is the suite for testing the
// InvalidRegistrationError type
type InvalidRegistrationErrorSuite struct {
	suite.Suite
}

// TestMessage tests the message names the index, both types and the call site
func (suite *InvalidRegistrationErrorSuite) TestMessage() {
	// Arrange
	err := &InvalidRegistrationError{
		Index:       3,
		ServiceType: reflect.TypeFor[testRepository](),
		ValueType:   reflect.TypeFor[string](),
		Site:        callSite{File: "/src/ordo/main.go", Line: 42},
		Err:         ErrFactoryNotFunction,
	}

	// Act
	msg := err.Error()

	// Assert
	suite.Contains(msg, "option 3")
	suite.Contains(msg, "main.go:42")
	suite.Contains(msg, "ordo.testRepository")
	suite.Contains(msg, "string")
	suite.Contains(msg, "service factory must be a function")
}

// TestMessageWithoutOptionalFields tests the message renders without a service
// type, a value type or a call site
func (suite *InvalidRegistrationErrorSuite) TestMessageWithoutOptionalFields() {
	// Arrange
	err := &InvalidRegistrationError{
		Index: 0,
		Err:   ErrNilRegistration,
	}

	// Act
	msg := err.Error()

	// Assert
	suite.Equal("option 0: registration value is nil", msg)
}

// TestUnwrapsCause tests the fault unwraps to its cause sentinel
func (suite *InvalidRegistrationErrorSuite) TestUnwrapsCause() {
	// Arrange
	err := &InvalidRegistrationError{
		Err: ErrNotAssignable,
	}

	// Act & Assert
	suite.ErrorIs(err, ErrNotAssignable)
	suite.NotErrorIs(err, ErrFactoryNotFunction)
	suite.Equal(ErrNotAssignable, errors.Unwrap(err))
}

// TestInvalidRegistrationError tests the InvalidRegistrationError type
func TestInvalidRegistrationError(t *testing.T) {
	suite.Run(t, new(InvalidRegistrationErrorSuite))
}

// RegistrationErrorSuite is the suite for testing the RegistrationError type
type RegistrationErrorSuite struct {
	suite.Suite
}

// TestMessage tests the aggregate renders a header and one line per fault
func (suite *RegistrationErrorSuite) TestMessage() {
	// Arrange
	err := &RegistrationError{
		Faults: []error{
			&InvalidRegistrationError{Index: 0, Err: ErrFactoryNotFunction},
			&InvalidRegistrationError{Index: 1, Err: ErrFactoryNoReturn},
		},
	}

	// Act
	msg := err.Error()

	// Assert
	suite.Contains(msg, "ordo: container registration failed:")
	suite.Contains(msg, "\n  - option 0: service factory must be a function")
	suite.Contains(msg, "\n  - option 1: service factory must return at least one value")
}

// TestTraversesEveryFault tests the standard helpers reach every fault
func (suite *RegistrationErrorSuite) TestTraversesEveryFault() {
	// Arrange
	err := &RegistrationError{
		Faults: []error{
			&InvalidRegistrationError{Index: 0, Err: ErrFactoryNotFunction},
			&InvalidRegistrationError{Index: 1, Err: ErrNotAssignable},
		},
	}

	// Act & Assert
	suite.ErrorIs(err, ErrFactoryNotFunction)
	suite.ErrorIs(err, ErrNotAssignable)
	suite.NotErrorIs(err, ErrNilRegistration)

	var invalid *InvalidRegistrationError
	suite.Require().ErrorAs(err, &invalid)
	suite.Equal(0, invalid.Index)
}

// TestRegistrationError tests the RegistrationError type
func TestRegistrationError(t *testing.T) {
	suite.Run(t, new(RegistrationErrorSuite))
}

// registrationFaultCase is a malformed registration and the cause it must report
type registrationFaultCase struct {
	// Name is the case name
	Name string

	// Opt is the malformed registration
	Opt Option

	// Cause is the sentinel the reported fault must unwrap to
	Cause error
}

// ContainerRegistrationSuite is the suite for testing the registration phase of
// the New function
type ContainerRegistrationSuite struct {
	suite.Suite
}

// TestFaultCoverage tests every malformed registration is reported as an error
func (suite *ContainerRegistrationSuite) TestFaultCoverage() {
	// Arrange
	cases := []registrationFaultCase{
		{
			Name:  "factory is not a function",
			Opt:   WithFactory("not a factory"),
			Cause: ErrFactoryNotFunction,
		},
		{
			Name:  "factory returns no value",
			Opt:   WithFactory(func() {}),
			Cause: ErrFactoryNoReturn,
		},
		{
			Name:  "factory returns too many values",
			Opt:   WithFactory(func() (int, error, int) { return 0, nil, 0 }),
			Cause: ErrFactoryTooManyReturns,
		},
		{
			Name:  "second return value is not an error",
			Opt:   WithFactory(func() (int, string) { return 0, "" }),
			Cause: ErrFactorySecondReturnNotErr,
		},
		{
			Name:  "factory return type is not assignable",
			Opt:   WithService[testRepository](func() *testService { return &testService{} }),
			Cause: ErrNotAssignable,
		},
		{
			Name:  "instance is not assignable",
			Opt:   WithService[testRepository](&testService{}),
			Cause: ErrNotAssignable,
		},
		{
			Name:  "nil registration value",
			Opt:   WithService[testRepository](nil),
			Cause: ErrNilRegistration,
		},
		{
			Name:  "nil keyed registration value",
			Opt:   WithKeyedService[testRepository]("db", nil),
			Cause: ErrNilRegistration,
		},
		{
			Name:  "nil factory",
			Opt:   WithFactory(nil),
			Cause: ErrNilRegistration,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.Name, func() {
			// Act
			c, err := New(tc.Opt)

			// Assert
			suite.Nil(c)
			suite.Require().Error(err)

			var regErr *RegistrationError
			suite.Require().ErrorAs(err, &regErr)
			suite.Len(regErr.Faults, 1)
			suite.ErrorIs(err, tc.Cause)
			suite.Contains(err.Error(), "ordo: container registration failed:")

			var invalid *InvalidRegistrationError
			suite.Require().ErrorAs(err, &invalid)
			suite.Equal(0, invalid.Index)
			suite.Equal(registrationTestFile, filepath.Base(invalid.Site.File))
		})
	}
}

// TestNilRegistrationDoesNotPanic tests a nil registration value is reported
// rather than dereferenced
func (suite *ContainerRegistrationSuite) TestNilRegistrationDoesNotPanic() {
	// Act & Assert
	suite.NotPanics(func() {
		c, err := New(WithService[testRepository](nil))

		suite.Nil(c)
		suite.ErrorIs(err, ErrNilRegistration)
	})
}

// TestReportsEveryFault tests every malformed registration of a single
// construction is reported, tagged with its own option index
func (suite *ContainerRegistrationSuite) TestReportsEveryFault() {
	// Act
	c, err := New(
		WithFactory("not a factory"),
		WithService[testRepository](newTestRepository),
		WithFactory(func() {}),
	)

	// Assert
	suite.Nil(c)
	suite.Require().Error(err)

	var regErr *RegistrationError
	suite.Require().ErrorAs(err, &regErr)
	suite.Require().Len(regErr.Faults, 2)

	indexes := make([]int, 0, len(regErr.Faults))
	for _, fault := range regErr.Faults {
		var invalid *InvalidRegistrationError
		suite.Require().ErrorAs(fault, &invalid)
		indexes = append(indexes, invalid.Index)
	}

	suite.Equal([]int{0, 2}, indexes)
	suite.ErrorIs(err, ErrFactoryNotFunction)
	suite.ErrorIs(err, ErrFactoryNoReturn)
}

// TestFaultsReportTheRegistrationSite tests a fault points at the registration
// site instead of the construction call
func (suite *ContainerRegistrationSuite) TestFaultsReportTheRegistrationSite() {
	// Arrange
	opt, optLine := badRegistrationFromHelper()

	// Act
	_, err := New(opt)
	_, _, containerLine, _ := runtime.Caller(0)

	// Assert
	var invalid *InvalidRegistrationError
	suite.Require().ErrorAs(err, &invalid)
	suite.Equal(registrationTestFile, filepath.Base(invalid.Site.File))
	suite.Equal(optLine, invalid.Site.Line)
	suite.NotEqual(containerLine-1, invalid.Site.Line)
	suite.Contains(err.Error(), invalid.Site.String())
}

// TestSameServiceTypeFaultsAreDistinguishable tests two malformed
// registrations of the same service type differ in index and site
func (suite *ContainerRegistrationSuite) TestSameServiceTypeFaultsAreDistinguishable() {
	// Arrange
	first := WithService[testRepository]("not a factory")
	second := WithService[testRepository](&testService{})

	// Act
	_, err := New(first, second)

	// Assert
	var regErr *RegistrationError
	suite.Require().ErrorAs(err, &regErr)
	suite.Require().Len(regErr.Faults, 2)

	var firstFault, secondFault *InvalidRegistrationError
	suite.Require().ErrorAs(regErr.Faults[0], &firstFault)
	suite.Require().ErrorAs(regErr.Faults[1], &secondFault)

	suite.Equal(0, firstFault.Index)
	suite.Equal(1, secondFault.Index)
	suite.NotEqual(firstFault.Site.Line, secondFault.Site.Line)
	suite.Equal(firstFault.ServiceType, secondFault.ServiceType)
}

// TestRegistrationFaultsSkipVerification tests a registration fault suppresses
// the verification faults its own dropped registration would cause
func (suite *ContainerRegistrationSuite) TestRegistrationFaultsSkipVerification() {
	// Act
	c, err := New(
		WithService[testRepository]("not a factory"),
		WithFactory(newTestService),
	)

	// Assert
	suite.Nil(c)
	suite.Require().Error(err)

	var regErr *RegistrationError
	suite.Require().ErrorAs(err, &regErr)
	suite.Len(regErr.Faults, 1)

	var verErr *VerificationError
	suite.False(errors.As(err, &verErr))

	var missing *MissingDependencyError
	suite.False(errors.As(err, &missing))
	suite.NotContains(err.Error(), "ordo: container verification failed:")
}

// TestFaultyGraphStillReportsVerification tests a sound registration set with a
// faulty graph is still reported by verification
func (suite *ContainerRegistrationSuite) TestFaultyGraphStillReportsVerification() {
	// Act
	c, err := New(WithFactory(newTestService))

	// Assert
	suite.Nil(c)
	suite.Require().Error(err)

	var verErr *VerificationError
	suite.Require().ErrorAs(err, &verErr)

	var regErr *RegistrationError
	suite.False(errors.As(err, &regErr))
}

// TestNoFactoryIsCalled tests registration validation invokes no factory
func (suite *ContainerRegistrationSuite) TestNoFactoryIsCalled() {
	// Arrange
	repoCalls, serviceCalls := 0, 0

	// Act
	c, err := New(
		WithService[testRepository](func() testRepository {
			repoCalls++
			return &testRepositoryImpl{}
		}),
		WithFactory(func(repo testRepository) *testService {
			serviceCalls++
			return &testService{repo}
		}),
		WithFactory("not a factory"),
	)

	// Assert
	suite.Nil(c)
	suite.Require().Error(err)
	suite.Equal(0, repoCalls)
	suite.Equal(0, serviceCalls)
}

// TestSoundRegistrationsConstruct tests well formed registrations still
// construct a usable Container
func (suite *ContainerRegistrationSuite) TestSoundRegistrationsConstruct() {
	// Act
	c, err := New(
		WithService[testRepository](newTestRepository),
		WithFactory(newTestService),
		WithKeyedValue[testLogger]("main", &testLoggerImpl{}),
	)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(c)

	service, err := c.GetService[*testService]()
	suite.NoError(err)
	suite.NotNil(service.Repo)

	logger, err := c.GetKeyedService[testLogger]("main")
	suite.NoError(err)
	suite.NotNil(logger)
}

// TestContainerRegistration tests the registration phase of New
func TestContainerRegistration(t *testing.T) {
	suite.Run(t, new(ContainerRegistrationSuite))
}
