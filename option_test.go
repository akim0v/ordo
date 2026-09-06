package ordo

import (
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
