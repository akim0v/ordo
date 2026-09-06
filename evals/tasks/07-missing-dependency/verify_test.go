package main

import "testing"

func TestServiceResolves(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	svc, err := c.GetService[*Service]()
	if err != nil {
		t.Fatalf("resolve *Service: %v", err)
	}

	if got := svc.Kind(); got != "mem" {
		t.Fatalf("Kind() = %q, want %q", got, "mem")
	}
}
