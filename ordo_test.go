package ordo

import (
	"bytes"
	"errors"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)
// NewSuite is the suite for testing the New function
type NewSuite struct {
	suite.Suite
}

// TestSoundGraph tests a sound graph yields a usable Container and no error
func (suite *NewSuite) TestSoundGraph() {
	// Arrange
	opts := []Option{
		WithService[testRepository](newTestRepository),
		WithFactory(newTestService),
		WithFactory(newTestController),
	}

	// Act
	c, err := New(opts...)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(c)

	controller, err := c.GetService[*testController]()
	suite.NoError(err)
	suite.NotNil(controller.Service.Repo)
}

// TestEmptyContainer tests a Container with no registrations is sound
func (suite *NewSuite) TestEmptyContainer() {
	// Act
	c, err := New()

	// Assert
	suite.NoError(err)
	suite.NotNil(c)
}

// TestFaultyGraphYieldsNilContainer tests a faulty graph yields a nil Container
// and a describing error
func (suite *NewSuite) TestFaultyGraphYieldsNilContainer() {
	// Act
	c, err := New(WithFactory(newTestService))

	// Assert
	suite.Nil(c)
	suite.Require().Error(err)

	var verErr *VerificationError
	suite.Require().ErrorAs(err, &verErr)
	suite.Len(verErr.Faults, 1)
	suite.Contains(err.Error(), "ordo: container verification failed:")
	suite.ErrorIs(err, ErrServiceNotFound)
}

// TestCycleDoesNotDeadlock tests a cycle is returned as an error instead of
// deadlocking the process
func (suite *NewSuite) TestCycleDoesNotDeadlock() {
	// Act
	c, err := New(
		WithFactory(func(b *testServiceB) *testServiceA { return &testServiceA{b} }),
		WithFactory(func(a *testServiceA) *testServiceB { return &testServiceB{} }),
	)

	// Assert
	suite.Nil(c)
	suite.Require().Error(err)

	var cycleErr *CircularDependencyError
	suite.Require().ErrorAs(err, &cycleErr)
	suite.Equal(
		[]reflect.Type{
			reflect.TypeFor[*testServiceA](),
			reflect.TypeFor[*testServiceB](),
		},
		cycleErr.Cycle,
	)
}

// TestFactoriesNotCalled tests no factory is invoked while the Container is
// constructed
func (suite *NewSuite) TestFactoriesNotCalled() {
	// Arrange
	repoCalls, serviceCalls, controllerCalls := 0, 0, 0

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
		WithFactory(func(service *testService) *testController {
			controllerCalls++
			return &testController{service}
		}),
	)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(c)
	suite.Equal(0, repoCalls)
	suite.Equal(0, serviceCalls)
	suite.Equal(0, controllerCalls)
}

// TestLazySingleton tests a service is created on first resolution and reused
// afterwards
func (suite *NewSuite) TestLazySingleton() {
	// Arrange
	calls := 0
	c, err := New(
		WithService[testRepository](func() testRepository {
			calls++
			return &testRepositoryImpl{}
		}),
	)
	suite.Require().NoError(err)

	// Act
	first, err1 := c.GetService[testRepository]()
	second, err2 := c.GetService[testRepository]()

	// Assert
	suite.Equal(1, calls)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.Same(first, second)
}

// TestFactoryErrorSurfacesOnResolution tests a factory error does not fail the
// construction and is returned on resolution
func (suite *NewSuite) TestFactoryErrorSurfacesOnResolution() {
	// Arrange
	calls := 0

	// Act
	c, err := New(
		WithService[testRepository](func() (testRepository, error) {
			calls++
			return nil, errors.ErrUnsupported
		}),
		WithFactory(newTestService),
	)

	// Assert
	suite.Require().NoError(err)
	suite.Require().NotNil(c)
	suite.Equal(0, calls)

	service, resErr := c.GetService[*testService]()
	suite.Nil(service)
	suite.Require().Error(resErr)
	suite.ErrorIs(resErr, errors.ErrUnsupported)
	suite.Equal(1, calls)
}

// TestNew tests the New function
func TestNew(t *testing.T) {
	suite.Run(t, new(NewSuite))
}

// GetServiceSuite is the suite for testing the GetService function
type GetServiceSuite struct {
	suite.Suite
}

// TestFactory tests the factory service
func (suite *GetServiceSuite) TestFactory() {
	// Arrange
	inst := "test"
	f := func() string {
		return inst
	}
	c, err := New(WithFactory(f))
	suite.Require().NoError(err)

	// Act
	res, err := c.GetService[string]()

	// Assert
	suite.Equal(inst, res)
	suite.NoError(err)
}

// TestFactoryError tests the factory service error
func (suite *GetServiceSuite) TestFactoryError() {
	// Arrange
	f := func() (string, error) {
		return "", errors.ErrUnsupported
	}
	c, err := New(WithFactory(f))
	suite.Require().NoError(err)

	// Act
	res, err := c.GetService[string]()

	// Assert
	suite.Empty(res)
	suite.Error(err)
}

// TestInstance tests the instance service
func (suite *GetServiceSuite) TestInstance() {
	// Arrange
	inst := "test"
	c, err := New(WithValue(inst))
	suite.Require().NoError(err)

	// Act
	res, err := c.GetService[string]()

	// Assert
	suite.Equal(inst, res)
	suite.NoError(err)
}

// TestGetService tests the GetService function
func TestGetService(t *testing.T) {
	suite.Run(t, new(GetServiceSuite))
}

// GetKeyedServiceSuite is the suite for testing the GetKeyedService function
type GetKeyedServiceSuite struct {
	suite.Suite
}

// TestInstance tests the factory service
func (suite *GetKeyedServiceSuite) TestFactory() {
	// Arrange
	key := "key"
	inst := "test"
	f := func() string {
		return inst
	}
	c, err := New(WithKeyedFactory(key, f))
	suite.Require().NoError(err)

	// Act
	res, err := c.GetKeyedService[string](key)

	// Assert
	suite.Equal(inst, res)
	suite.NoError(err)
}

// TestInstance tests the factory service error
func (suite *GetKeyedServiceSuite) TestFactoryError() {
	// Arrange
	key := "key"
	f := func() (string, error) {
		return "", errors.ErrUnsupported
	}
	c, err := New(WithKeyedFactory(key, f))
	suite.Require().NoError(err)

	// Act
	res, err := c.GetKeyedService[string](key)

	// Assert
	suite.Empty(res)
	suite.Error(err)
}

// TestInstance tests the instance service
func (suite *GetKeyedServiceSuite) TestInstance() {
	// Arrange
	key := "key"
	inst := "test"
	c, err := New(WithKeyedValue(key, inst))
	suite.Require().NoError(err)

	// Act
	res, err := c.GetKeyedService[string](key)

	// Assert
	suite.Equal(inst, res)
	suite.NoError(err)
}

// TestGetKeyedService tests the GetKeyedService function
func TestGetKeyedService(t *testing.T) {
	suite.Run(t, new(GetKeyedServiceSuite))
}

// MustGetServiceSuite is the suite for testing the MustGetService function
type MustGetServiceSuite struct {
	suite.Suite
}

// TestFactory tests the factory service
func (suite *MustGetServiceSuite) TestFactory() {
	// Arrange
	inst := "test"
	f := func() string {
		return inst
	}
	c, err := New(WithFactory(f))
	suite.Require().NoError(err)

	// Act & Assert
	var res string
	suite.NotPanics(func() {
		res = c.MustGetService[string]()
	})
	suite.Equal(inst, res)
}

// TestFactoryError tests the factory service error
func (suite *MustGetServiceSuite) TestFactoryError() {
	// Arrange
	f := func() (string, error) {
		return "", errors.ErrUnsupported
	}
	c, err := New(WithFactory(f))
	suite.Require().NoError(err)

	// Act & Assert
	suite.Panics(func() {
		c.MustGetService[string]()
	})
}

// TestInstance tests the instance service
func (suite *MustGetServiceSuite) TestInstance() {
	// Arrange
	inst := "test"
	c, err := New(WithValue(inst))
	suite.Require().NoError(err)

	// Act & Assert
	var res string
	suite.NotPanics(func() {
		res = c.MustGetService[string]()
	})
	suite.Equal(inst, res)
}

// TestMustGetService tests the MustGetService function
func TestMustGetService(t *testing.T) {
	suite.Run(t, new(MustGetServiceSuite))
}

// MustGetKeyedServiceSuite is the suite for testing the MustGetKeyedService function
type MustGetKeyedServiceSuite struct {
	suite.Suite
}

// TestFactory tests the factory service
func (suite *MustGetKeyedServiceSuite) TestFactory() {
	// Arrange
	key := "key"
	inst := "test"
	f := func() string {
		return inst
	}
	c, err := New(WithKeyedFactory(key, f))
	suite.Require().NoError(err)

	// Act & Assert
	var res string
	suite.NotPanics(func() {
		res = c.MustGetKeyedService[string](key)
	})
	suite.Equal(inst, res)
}

// TestFactoryError tests the factory service error
func (suite *MustGetKeyedServiceSuite) TestFactoryError() {
	// Arrange
	key := "key"
	f := func() (string, error) {
		return "", errors.ErrUnsupported
	}
	c, err := New(WithKeyedFactory(key, f))
	suite.Require().NoError(err)

	// Act & Assert
	suite.Panics(func() {
		c.MustGetKeyedService[string](key)
	})
}

// TestInstance tests the instance service
func (suite *MustGetKeyedServiceSuite) TestInstance() {
	// Arrange
	key := "key"
	inst := "test"
	c, err := New(WithKeyedValue(key, inst))
	suite.Require().NoError(err)

	// Act & Assert
	var res string
	suite.NotPanics(func() {
		res = c.MustGetKeyedService[string](key)
	})
	suite.Equal(inst, res)
}

// TestMustGetKeyedService tests the MustGetKeyedService function
func TestMustGetKeyedService(t *testing.T) {
	suite.Run(t, new(MustGetKeyedServiceSuite))
}

// VerificationAgreementSuite is the suite for testing the verify and getService
// methods reach the same verdict, so the two paths cannot drift
type VerificationAgreementSuite struct {
	suite.Suite
}

// agreementCase is a fixture graph with the dependency type under test
type agreementCase struct {
	// Name is the fixture name
	Name string

	// Opts are the fixture registrations
	Opts []Option

	// DepType is the dependency type both paths are asked about
	DepType reflect.Type
}

// reportsMissing reports whether verification found the dependency missing
func (suite *VerificationAgreementSuite) reportsMissing(c *Container, depType reflect.Type) bool {
	err := c.verify()
	if err == nil {
		return false
	}

	var verErr *VerificationError
	suite.Require().ErrorAs(err, &verErr)

	for _, fault := range verErr.Faults {
		if missing, ok := fault.(*MissingDependencyError); ok && missing.DependencyType == depType {
			return true
		}
	}

	return false
}

// TestAgreement tests both paths reach the same verdict for every fixture
func (suite *VerificationAgreementSuite) TestAgreement() {
	// Arrange
	cases := []agreementCase{
		{
			Name: "registered dependency",
			Opts: []Option{
				WithService[testRepository](newTestRepository),
				WithFactory(newTestService),
			},
			DepType: reflect.TypeFor[testRepository](),
		},
		{
			Name:    "unregistered dependency",
			Opts:    []Option{WithFactory(newTestService)},
			DepType: reflect.TypeFor[testRepository](),
		},
		{
			Name: "keyed only registration",
			Opts: []Option{
				WithKeyedService[testRepository]("db", newTestRepository),
				WithFactory(newTestService),
			},
			DepType: reflect.TypeFor[testRepository](),
		},
		{
			Name: "instance registration",
			Opts: []Option{
				WithValue[testRepository](&testRepositoryImpl{}),
				WithFactory(newTestService),
			},
			DepType: reflect.TypeFor[testRepository](),
		},
		{
			Name: "slice dependency with registrations",
			Opts: []Option{
				WithService[testRepository](newTestRepository),
				WithFactory(newTestCollector),
			},
			DepType: reflect.TypeFor[[]testRepository](),
		},
		{
			Name:    "slice dependency without registrations",
			Opts:    []Option{WithFactory(newTestCollector)},
			DepType: reflect.TypeFor[[]testRepository](),
		},
		{
			Name: "container dependency",
			Opts: []Option{
				WithFactory(func(cont *Container) *testService { return &testService{} }),
			},
			DepType: reflect.TypeFor[*Container](),
		},
	}

	for _, tc := range cases {
		suite.Run(tc.Name, func() {
			// Arrange
			// newContainer skips verification, so the resolution attempt runs
			// even on a graph New would reject
			c := newFixtureContainer(tc.Opts...)

			// Act
			verificationMissing := suite.reportsMissing(c, tc.DepType)

			_, resErr := c.getService(newDependencyIdentifier(tc.DepType))
			resolutionMissing := errors.Is(resErr, ErrServiceNotFound)

			// Assert
			suite.Equal(
				resolutionMissing,
				verificationMissing,
				"verification and resolution disagree on %v",
				tc.DepType,
			)
		})
	}
}

// TestVerificationAgreement tests the verify and getService methods agree
func TestVerificationAgreement(t *testing.T) {
	suite.Run(t, new(VerificationAgreementSuite))
}

// ResolutionErrorSuite is the suite for testing the failure contract of the
// error returning resolution methods
type ResolutionErrorSuite struct {
	suite.Suite
}

// TestUnregisteredInterface tests an unregistered interface service is reported
// as an error instead of panicking on the type assertion
func (suite *ResolutionErrorSuite) TestUnregisteredInterface() {
	// Arrange
	c, err := New()
	suite.Require().NoError(err)

	// Act & Assert
	suite.NotPanics(func() {
		res, resErr := c.GetService[testRepository]()

		suite.Nil(res)
		suite.ErrorIs(resErr, ErrServiceNotFound)
	})
}

// TestUnregisteredKeyedInterface tests a keyed interface miss is reported as an
// error instead of panicking on the type assertion
func (suite *ResolutionErrorSuite) TestUnregisteredKeyedInterface() {
	// Arrange
	c, err := New(WithService[testRepository](newTestRepository))
	suite.Require().NoError(err)

	// Act & Assert
	suite.NotPanics(func() {
		res, resErr := c.GetKeyedService[testRepository]("missing")

		suite.Nil(res)
		suite.ErrorIs(resErr, ErrServiceNotFound)
	})
}

// TestInterfaceFactoryError tests an interface service whose factory fails is
// reported as an error instead of panicking on the type assertion
func (suite *ResolutionErrorSuite) TestInterfaceFactoryError() {
	// Arrange
	c, err := New(
		WithService[testRepository](func() (testRepository, error) {
			return nil, errors.ErrUnsupported
		}),
	)
	suite.Require().NoError(err)

	// Act & Assert
	suite.NotPanics(func() {
		res, resErr := c.GetService[testRepository]()

		suite.Nil(res)
		suite.ErrorIs(resErr, errors.ErrUnsupported)
	})
}

// TestUnregisteredInterfaceSlice tests an unregistered slice of an interface is
// reported as an error instead of panicking on the type assertion
func (suite *ResolutionErrorSuite) TestUnregisteredInterfaceSlice() {
	// Arrange
	c, err := New()
	suite.Require().NoError(err)

	// Act & Assert
	suite.NotPanics(func() {
		res, resErr := c.GetService[[]testRepository]()

		suite.Nil(res)
		suite.ErrorIs(resErr, ErrServiceNotFound)
	})
}

// TestRegisteredInterfaceResolves tests a registered interface service still
// resolves with no error
func (suite *ResolutionErrorSuite) TestRegisteredInterfaceResolves() {
	// Arrange
	c, err := New(WithService[testRepository](newTestRepository))
	suite.Require().NoError(err)

	// Act
	res, resErr := c.GetService[testRepository]()

	// Assert
	suite.NoError(resErr)
	suite.NotNil(res)
	suite.Equal("value", res.Get())
}

// TestResolutionErrors tests the failure contract of the resolution methods
func TestResolutionErrors(t *testing.T) {
	suite.Run(t, new(ResolutionErrorSuite))
}

// MustPanicPayloadSuite is the suite for testing what the must style accessors
// panic with
type MustPanicPayloadSuite struct {
	suite.Suite

	// logOutput collects everything written to the standard logger during a test
	logOutput bytes.Buffer
}

// SetupTest redirects the standard logger output to the suite buffer
func (suite *MustPanicPayloadSuite) SetupTest() {
	suite.logOutput.Reset()
	log.SetOutput(&suite.logOutput)
}

// TearDownTest restores the standard logger output
func (suite *MustPanicPayloadSuite) TearDownTest() {
	log.SetOutput(os.Stderr)
}

// recoverPanic returns the value the provided function panicked with
func (suite *MustPanicPayloadSuite) recoverPanic(f func()) (recovered any) {
	defer func() {
		recovered = recover()
	}()

	f()

	return nil
}

// TestPanicsWithResolutionError tests the panic value is the error the error
// returning resolution reports
func (suite *MustPanicPayloadSuite) TestPanicsWithResolutionError() {
	// Arrange
	c, err := New(
		WithService[testRepository](func() (testRepository, error) {
			return nil, errors.ErrUnsupported
		}),
	)
	suite.Require().NoError(err)

	_, expected := c.GetService[testRepository]()
	suite.Require().Error(expected)

	// Act
	recovered := suite.recoverPanic(func() {
		c.MustGetService[testRepository]()
	})

	// Assert
	recoveredErr, ok := recovered.(error)
	suite.Require().True(ok, "recovered value must be an error, got %T", recovered)
	suite.Equal(expected, recoveredErr)
	suite.ErrorIs(recoveredErr, errors.ErrUnsupported)
	suite.Empty(suite.logOutput.String())
}

// TestKeyedPanicsWithResolutionError tests the keyed accessor panics with the
// error the error returning resolution reports
func (suite *MustPanicPayloadSuite) TestKeyedPanicsWithResolutionError() {
	// Arrange
	c, err := New(WithService[testRepository](newTestRepository))
	suite.Require().NoError(err)

	_, expected := c.GetKeyedService[testRepository]("missing")
	suite.Require().Error(expected)

	// Act
	recovered := suite.recoverPanic(func() {
		c.MustGetKeyedService[testRepository]("missing")
	})

	// Assert
	recoveredErr, ok := recovered.(error)
	suite.Require().True(ok, "recovered value must be an error, got %T", recovered)
	suite.ErrorIs(recoveredErr, ErrServiceNotFound)
	suite.Empty(suite.logOutput.String())
}

// TestResolvedServiceDoesNotPanic tests a resolvable service is returned
// without panicking and without log output
func (suite *MustPanicPayloadSuite) TestResolvedServiceDoesNotPanic() {
	// Arrange
	c, err := New(WithService[testRepository](newTestRepository))
	suite.Require().NoError(err)

	// Act
	recovered := suite.recoverPanic(func() {
		suite.NotNil(c.MustGetService[testRepository]())
	})

	// Assert
	suite.Nil(recovered)
	suite.Empty(suite.logOutput.String())
}

// TestMustPanicPayload tests what the must style accessors panic with
func TestMustPanicPayload(t *testing.T) {
	suite.Run(t, new(MustPanicPayloadSuite))
}
