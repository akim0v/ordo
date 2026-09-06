// Example: verification
//
// This is the guarantee ordo is built around: a container either fails to
// build, or is fully resolvable. Once every registration is well formed, the
// container walks the whole dependency graph before returning, so a missing
// dependency or a cycle is reported at NewContainer instead of surfacing on the
// first resolution in production.
package main

import (
	"errors"
	"fmt"

	"github.com/akim0v/ordo/di"
)

type UserRepository interface {
	GetUserName(id int) (string, bool)
}

type UserService struct{}

// NewUserService needs a UserRepository that the first container below never
// registers.
func NewUserService(UserRepository) *UserService { return &UserService{} }

// Billing and Accounts depend on each other, forming a cycle.
type Billing struct{}
type Accounts struct{}

func NewBilling(*Accounts) *Billing  { return &Billing{} }
func NewAccounts(*Billing) *Accounts { return &Accounts{} }

func main() {
	missingDependency()
	fmt.Println()
	circularDependency()
}

func missingDependency() {
	fmt.Println("=== missing dependency ===")

	// UserService is registered, its UserRepository dependency is not.
	_, err := di.NewContainer(
		di.WithFactory(NewUserService),
	)

	fmt.Println(err)
	// di: container verification failed:
	//   - service "*main.UserService" requires "main.UserRepository", which is not registered

	// A missing dependency wraps the same sentinel a failed runtime resolution
	// would return.
	fmt.Println(errors.Is(err, di.ErrServiceNotFound)) // true

	// Every fault of the pass is aggregated and reachable with errors.As.
	var verification *di.VerificationError
	if errors.As(err, &verification) {
		fmt.Printf("faults: %d\n", len(verification.Faults))
	}

	var missing *di.MissingDependencyError
	if errors.As(err, &missing) {
		fmt.Printf("%s is missing %s\n", missing.RequestingType, missing.DependencyType)
	}
}

func circularDependency() {
	fmt.Println("=== circular dependency ===")

	// Billing needs Accounts, Accounts needs Billing. Neither can ever be built.
	_, err := di.NewContainer(
		di.WithFactory(NewBilling),
		di.WithFactory(NewAccounts),
	)

	fmt.Println(err)

	// The fault carries the ordered cycle, so the offending edge is readable
	// rather than guessed at.
	var cycle *di.CircularDependencyError
	if errors.As(err, &cycle) {
		for i, typ := range cycle.Cycle {
			fmt.Printf("  %d. %s\n", i+1, typ)
		}
	}
}
