package ordo

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// WithServiceSuite is the suite for testing the WithService function
type WithServiceSuite struct {
	suite.Suite
}

// TestInstance tests the instance service
func (suite *WithServiceSuite) TestInstance() {
	// Arrange
	inst := "test"

	opt := WithService[string](inst)
	c, err := New(opt)
	suite.Require().NoError(err)

	id := serviceIdentifier{
		Type: reflect.TypeOf(inst),
	}

	// Act
	lastAccessor := c.accessors[id].Last()

	// Assert
	suite.Equal(id, lastAccessor.id)
	suite.Equal(c, lastAccessor.cont)
	suite.Nil(lastAccessor.factory)
	suite.Equal(inst, lastAccessor.instance.Interface())
	suite.NoError(lastAccessor.err)
}

// TestInstance tests the instance service
func (suite *WithServiceSuite) TestFactory() {
	// Arrange
	inst := "test"
	f := func() string {
		return inst
	}

	opt := WithService[string](f)
	c, err := New(opt)
	suite.Require().NoError(err)

	id := serviceIdentifier{
		Type: reflect.TypeOf(inst),
	}

	// Act
	lastAccessor := c.accessors[id].Last()

	// Assert
	suite.Equal(id, lastAccessor.id)
	suite.Equal(c, lastAccessor.cont)
	suite.NotNil(lastAccessor.factory)
	suite.Nil(lastAccessor.instance)
	suite.NoError(lastAccessor.err)
}

// TestWithService tests the WithService function
func TestWithService(t *testing.T) {
	suite.Run(t, new(WithServiceSuite))
}

// WithKeyedServiceSuite is the suite for testing the WithKeyedService function
type WithKeyedServiceSuite struct {
	suite.Suite
}

// TestInstance tests the instance service
func (suite *WithKeyedServiceSuite) TestInstance() {
	// Arrange
	key := "key"
	inst := "test"

	opt := WithKeyedService[string](key, inst)
	c, err := New(opt)
	suite.Require().NoError(err)

	id := serviceIdentifier{
		Type:   reflect.TypeOf(inst),
		Key:    key,
		HasKey: true,
	}

	// Act
	lastAccessor := c.accessors[id].Last()

	// Assert
	suite.Equal(id, lastAccessor.id)
	suite.Equal(c, lastAccessor.cont)
	suite.Nil(lastAccessor.factory)
	suite.Equal(inst, lastAccessor.instance.Interface())
	suite.NoError(lastAccessor.err)
}

// TestInstance tests the instance service
func (suite *WithKeyedServiceSuite) TestFactory() {
	// Arrange
	key := "key"
	inst := "test"
	f := func() string {
		return inst
	}

	opt := WithKeyedService[string](key, f)
	c, err := New(opt)
	suite.Require().NoError(err)

	id := serviceIdentifier{
		Type:   reflect.TypeOf(inst),
		Key:    key,
		HasKey: true,
	}

	// Act
	lastAccessor := c.accessors[id].Last()

	// Assert
	suite.Equal(id, lastAccessor.id)
	suite.Equal(c, lastAccessor.cont)
	suite.NotNil(lastAccessor.factory)
	suite.Nil(lastAccessor.instance)
	suite.NoError(lastAccessor.err)
}

// TestWithKeyedService tests the WithKeyedService function
func TestWithKeyedService(t *testing.T) {
	suite.Run(t, new(WithKeyedServiceSuite))
}

// TestWithValue tests the WithValue function
func TestWithValue(t *testing.T) {
	// Arrange
	inst := "test"

	opt := WithValue(inst)
	c, err := New(opt)
	require.NoError(t, err)

	id := serviceIdentifier{
		Type: reflect.TypeOf(inst),
	}

	// Act
	lastAccessor := c.accessors[id].Last()

	// Assert
	assert.Equal(t, id, lastAccessor.id)
	assert.Equal(t, c, lastAccessor.cont)
	assert.Nil(t, lastAccessor.factory)
	assert.Equal(t, inst, lastAccessor.instance.Interface())
	assert.NoError(t, lastAccessor.err)
}

// TestWithKeyedValue tests the WithKeyedValue function
func TestWithKeyedValue(t *testing.T) {
	// Arrange
	key := "key"
	inst := "test"

	opt := WithKeyedValue[string](key, inst)
	c, err := New(opt)
	require.NoError(t, err)

	id := serviceIdentifier{
		Type:   reflect.TypeOf(inst),
		Key:    key,
		HasKey: true,
	}

	// Act
	lastAccessor := c.accessors[id].Last()

	// Assert
	assert.Equal(t, id, lastAccessor.id)
	assert.Equal(t, c, lastAccessor.cont)
	assert.Nil(t, lastAccessor.factory)
	assert.Equal(t, inst, lastAccessor.instance.Interface())
	assert.NoError(t, lastAccessor.err)
}

// TestWithFactory tests the WithFactory function
func TestWithFactory(t *testing.T) {
	// Arrange
	inst := "test"
	f := func() string {
		return inst
	}

	opt := WithFactory(f)
	c, err := New(opt)
	require.NoError(t, err)

	id := serviceIdentifier{
		Type: reflect.TypeOf(inst),
	}

	// Act
	lastAccessor := c.accessors[id].Last()
	res, err := lastAccessor.factory.Call()

	// Assert
	assert.Equal(t, id, lastAccessor.id)
	assert.Equal(t, c, lastAccessor.cont)
	assert.Nil(t, lastAccessor.instance)
	assert.NotNil(t, lastAccessor.factory)
	assert.NoError(t, lastAccessor.err)

	assert.Equal(t, inst, res.Interface())
	assert.NoError(t, err)
}

// TestWithKeyedFactory tests the WithKeyedFactory function
func TestWithKeyedFactory(t *testing.T) {
	// Arrange
	key := "key"
	inst := "test"
	f := func() string {
		return inst
	}

	opt := WithKeyedFactory(key, f)
	c, err := New(opt)
	require.NoError(t, err)

	id := serviceIdentifier{
		Type:   reflect.TypeOf(inst),
		Key:    key,
		HasKey: true,
	}

	// Act
	lastAccessor := c.accessors[id].Last()
	res, err := lastAccessor.factory.Call()

	// Assert
	assert.Equal(t, id, lastAccessor.id)
	assert.Equal(t, c, lastAccessor.cont)
	assert.Nil(t, lastAccessor.instance)
	assert.NotNil(t, lastAccessor.factory)
	assert.NoError(t, lastAccessor.err)

	assert.Equal(t, inst, res.Interface())
	assert.NoError(t, err)
}

// TestMultiple tests the adding multiple services with the same identifier
func TestMultiple(t *testing.T) {
	// Arrange
	inst1, inst2 := "test1", "test2"
	opt1, opt2 := WithValue(inst1), WithValue(inst2)

	id := serviceIdentifier{
		Type: reflect.TypeFor[string](),
	}

	// Act
	c, err := New(opt1, opt2)
	require.NoError(t, err)

	res := make([]*serviceAccessor, 0, 2)
	for _, a := range c.accessors[id].Iter() {
		res = append(res, a)
	}

	// Assert
	if assert.Equal(t, 2, len(res)) {
		assert.Equal(t, inst2, res[1].instance.Interface())
		assert.Equal(t, inst1, res[0].instance.Interface())
	}
}

// ValueIsFunctionSuite is the suite for the fault reported when a constructor
// is registered through an option that infers the service type from its
// argument.
type ValueIsFunctionSuite struct {
	suite.Suite
}

// TestConstructorThroughWithValue tests a constructor passed to WithValue is
// reported as a function registered as a value.
//
// The service type is inferred as the function's own type, so the generic
// ErrNotAssignable message would name the func type as the service and the
// return type as the offending value, which reads backwards.
func (suite *ValueIsFunctionSuite) TestConstructorThroughWithValue() {
	// Arrange
	newRepository := func() *testRepositoryImpl { return &testRepositoryImpl{} }

	// Act
	c, err := New(WithValue(newRepository))

	// Assert
	suite.Require().Error(err)
	suite.Nil(c)
	suite.ErrorIs(err, ErrValueIsFunction)
	suite.NotErrorIs(err, ErrNotAssignable)

	invalid, ok := errors.AsType[*InvalidRegistrationError](err)
	suite.Require().True(ok)
	suite.Equal(0, invalid.Index)
	suite.Equal(reflect.TypeOf(newRepository), invalid.ServiceType)

	// The value type is omitted: naming it would repeat the service type
	suite.Nil(invalid.ValueType)
}

// TestKeyedValueReportsTheSameFault tests the keyed form is reported the same
// way, since it infers the service type from its argument too.
func (suite *ValueIsFunctionSuite) TestKeyedValueReportsTheSameFault() {
	// Arrange
	newRepository := func() *testRepositoryImpl { return &testRepositoryImpl{} }

	// Act
	_, err := New(WithKeyedValue("primary", newRepository))

	// Assert
	suite.ErrorIs(err, ErrValueIsFunction)
}

// TestMalformedFactoryKeepsItsOwnCause tests the fault is only reported once
// the function is a usable factory shape.
//
// A function that cannot be a factory at all is rejected earlier and keeps the
// cause describing why, rather than being reported as a value registration.
func (suite *ValueIsFunctionSuite) TestMalformedFactoryKeepsItsOwnCause() {
	// Arrange
	returnsNothing := func() {}

	// Act
	_, err := New(WithValue(returnsNothing))

	// Assert
	suite.ErrorIs(err, ErrFactoryNoReturn)
	suite.NotErrorIs(err, ErrValueIsFunction)
}

// TestExplicitServiceTypeIsUnaffected tests a constructor registered with an
// explicit service type still reports an unassignable value, since the service
// type was not inferred from the argument.
func (suite *ValueIsFunctionSuite) TestExplicitServiceTypeIsUnaffected() {
	// Arrange
	notARepository := func() *testLoggerImpl { return &testLoggerImpl{} }

	// Act
	_, err := New(WithService[testRepository](notARepository))

	// Assert
	suite.ErrorIs(err, ErrNotAssignable)
	suite.NotErrorIs(err, ErrValueIsFunction)
}

func TestValueIsFunction(t *testing.T) {
	suite.Run(t, new(ValueIsFunctionSuite))
}
