package main

import "testing"

func TestServiceResolvesWithConfig(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	svc, err := c.GetService[*Service]()
	if err != nil {
		t.Fatalf("resolve *Service: %v", err)
	}

	if got := svc.Describe(); got != "app:mem" {
		t.Fatalf("Describe() = %q, want %q", got, "app:mem")
	}
}
