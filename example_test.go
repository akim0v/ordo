package ordo_test

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/akim0v/ordo"
)

// registeredAt matches the source location a verification fault carries.
var registeredAt = regexp.MustCompile(` registered at \S+`)

// withoutSite removes the "registered at file.go:12" a fault carries, so an
// example's expected output does not depend on its own line numbers. Real
// messages keep the location.
func withoutSite(err error) string {
	return registeredAt.ReplaceAllString(err.Error(), "")
}

// Config is a settings value the application builds itself.
type Config struct {
	DSN string
}

// UserRepository is the interface the service layer depends on.
type UserRepository interface {
	Name() string
}

// PostgresRepository is a UserRepository built from a *Config.
type PostgresRepository struct {
	dsn string
}

func NewPostgresRepository(cfg *Config) *PostgresRepository {
	return &PostgresRepository{dsn: cfg.DSN}
}

func (r *PostgresRepository) Name() string { return "postgres(" + r.dsn + ")" }

// CacheRepository is a second UserRepository implementation.
type CacheRepository struct{}

func NewCacheRepository() *CacheRepository { return &CacheRepository{} }

func (*CacheRepository) Name() string { return "cache" }

// UserService depends on a UserRepository and nothing else.
type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) RepositoryName() string { return s.repository.Name() }

// ReportService depends on every registered UserRepository.
type ReportService struct {
	repositories []UserRepository
}

func NewReportService(repositories []UserRepository) *ReportService {
	return &ReportService{repositories: repositories}
}

func (s *ReportService) Count() int { return len(s.repositories) }

// Example wires a small application: a value, an implementation bound to the
// interface above it, and a constructor whose parameter the container supplies.
func Example() {
	c, err := ordo.New(
		ordo.WithValue(&Config{DSN: "localhost"}),
		ordo.WithService[UserRepository](NewPostgresRepository),
		ordo.WithFactory(NewUserService),
	)
	if err != nil {
		panic(err)
	}

	service := c.MustGetService[*UserService]()
	fmt.Println(service.RepositoryName())

	// Output:
	// postgres(localhost)
}

// ExampleNew builds a container. Registrations are validated and the dependency
// graph is verified before New returns, so a container that is returned without
// an error is fully resolvable.
func ExampleNew() {
	c, err := ordo.New(
		ordo.WithService[UserRepository](NewCacheRepository),
		ordo.WithFactory(NewUserService),
	)

	fmt.Println(err)
	fmt.Println(c.MustGetService[*UserService]().RepositoryName())

	// Output:
	// <nil>
	// cache
}

// ExampleNew_missingDependency shows a dependency that is never registered. The
// fault is reported by New, not by the resolution that would have needed it.
func ExampleNew_missingDependency() {
	_, err := ordo.New(
		ordo.WithFactory(NewUserService),
	)

	// The real message also names the registration's source location.
	fmt.Println(withoutSite(err))
	fmt.Println(errors.Is(err, ordo.ErrServiceNotFound))

	// Output:
	// ordo: container verification failed:
	//   - service "*ordo_test.UserService" requires "ordo_test.UserRepository", which is not registered
	// true
}

// ExampleNew_circularDependency shows two services that depend on each other.
// Neither could ever be built, and New says so instead of deadlocking or
// overflowing the stack on first use.
func ExampleNew_circularDependency() {
	type Billing struct{}
	type Accounts struct{}

	_, err := ordo.New(
		ordo.WithFactory(func(*Accounts) *Billing { return nil }),
		ordo.WithFactory(func(*Billing) *Accounts { return nil }),
	)

	cycle, ok := errors.AsType[*ordo.CircularDependencyError](err)
	fmt.Println(ok)
	fmt.Println(len(cycle.Cycle))

	// Output:
	// true
	// 2
}

// ExampleNew_registrationFault shows a malformed option. Registration is checked
// before the graph, and a rejected option means the graph is never verified.
func ExampleNew_registrationFault() {
	_, err := ordo.New(
		ordo.WithFactory("not a function"),
	)

	_, ok := errors.AsType[*ordo.RegistrationError](err)
	fmt.Println(ok)
	fmt.Println(errors.Is(err, ordo.ErrFactoryNotFunction))

	// Output:
	// true
	// true
}

// ExampleWithService binds a constructor to the interface it satisfies, so the
// service is resolvable as UserRepository rather than as its concrete type.
func ExampleWithService() {
	c, _ := ordo.New(
		ordo.WithService[UserRepository](NewCacheRepository),
	)

	fmt.Println(c.MustGetService[UserRepository]().Name())

	// Output:
	// cache
}

// ExampleWithService_instance registers an already-built value. WithService
// accepts a constructor or an instance and tells them apart by kind.
func ExampleWithService_instance() {
	c, _ := ordo.New(
		ordo.WithService[UserRepository](&CacheRepository{}),
	)

	fmt.Println(c.MustGetService[UserRepository]().Name())

	// Output:
	// cache
}

// ExampleWithFactory registers a constructor under its own return type, without
// naming that type at the call site.
func ExampleWithFactory() {
	c, _ := ordo.New(
		ordo.WithService[UserRepository](NewCacheRepository),
		ordo.WithFactory(NewUserService),
	)

	fmt.Println(c.MustGetService[*UserService]().RepositoryName())

	// Output:
	// cache
}

// ExampleWithFactory_error shows the second supported constructor shape. A
// constructor may return (T, error); the error surfaces when the service is
// resolved, since the graph itself is sound.
func ExampleWithFactory_error() {
	failing := errors.New("connection refused")

	c, err := ordo.New(
		ordo.WithFactory(func() (*CacheRepository, error) { return nil, failing }),
	)
	fmt.Println(err)

	_, err = c.GetService[*CacheRepository]()
	fmt.Println(errors.Is(err, failing))

	// Output:
	// <nil>
	// true
}

// ExampleWithValue registers a ready value under its own type.
func ExampleWithValue() {
	c, _ := ordo.New(
		ordo.WithValue(&Config{DSN: "localhost"}),
	)

	fmt.Println(c.MustGetService[*Config]().DSN)

	// Output:
	// localhost
}

// ExampleWithKeyedService tells two registrations of one type apart by name.
//
// Keys are a call-site facility: a factory parameter is always resolved without
// a key, so a keyed registration never satisfies a constructor.
func ExampleWithKeyedService() {
	c, _ := ordo.New(
		ordo.WithKeyedService[UserRepository]("cache", NewCacheRepository),
		ordo.WithKeyedService[UserRepository]("postgres", func() *PostgresRepository {
			return &PostgresRepository{dsn: "localhost"}
		}),
	)

	fmt.Println(c.MustGetKeyedService[UserRepository]("cache").Name())
	fmt.Println(c.MustGetKeyedService[UserRepository]("postgres").Name())

	// Output:
	// cache
	// postgres(localhost)
}

// ExampleWithKeyedFactory registers a constructor under its return type and a key.
func ExampleWithKeyedFactory() {
	c, _ := ordo.New(
		ordo.WithKeyedFactory("primary", NewCacheRepository),
	)

	fmt.Println(c.MustGetKeyedService[*CacheRepository]("primary").Name())

	// Output:
	// cache
}

// ExampleWithKeyedValue registers a ready value under its own type and a key.
func ExampleWithKeyedValue() {
	c, _ := ordo.New(
		ordo.WithKeyedValue("primary", &Config{DSN: "localhost"}),
		ordo.WithKeyedValue("replica", &Config{DSN: "replica"}),
	)

	fmt.Println(c.MustGetKeyedService[*Config]("replica").DSN)

	// Output:
	// replica
}
