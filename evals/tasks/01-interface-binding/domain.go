package main

// Greeter is the interface callers depend on.
type Greeter interface {
	Greet() string
}

type englishGreeter struct{}

// NewEnglishGreeter constructs the only implementation of Greeter.
func NewEnglishGreeter() *englishGreeter { return &englishGreeter{} }

func (*englishGreeter) Greet() string { return "hello" }

func main() {}
