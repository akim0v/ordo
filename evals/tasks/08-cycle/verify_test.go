package main

import "testing"

func TestAResolves(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	if _, err := c.GetService[*A](); err != nil {
		t.Fatalf("resolve *A: %v", err)
	}
}
