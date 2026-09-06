package main

import "testing"

func TestGreeterResolves(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	g, err := c.GetService[Greeter]()
	if err != nil {
		t.Fatalf("resolve Greeter: %v", err)
	}

	if got := g.Greet(); got != "hello" {
		t.Fatalf("Greet() = %q, want %q", got, "hello")
	}
}
