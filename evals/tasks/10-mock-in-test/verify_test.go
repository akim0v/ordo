package main

import "testing"

type fakeStore struct{}

func (*fakeStore) Kind() string { return "fake" }

func TestRealStoreIsWiredByDefault(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	svc, err := c.GetService[*Service]()
	if err != nil {
		t.Fatalf("resolve *Service: %v", err)
	}

	if got := svc.Kind(); got != "real" {
		t.Fatalf("Kind() = %q, want %q", got, "real")
	}
}

func TestMockIsSubstituted(t *testing.T) {
	c, err := BuildContainerWith(&fakeStore{})
	if err != nil {
		t.Fatalf("BuildContainerWith: %v", err)
	}

	svc, err := c.GetService[*Service]()
	if err != nil {
		t.Fatalf("resolve *Service: %v", err)
	}

	if got := svc.Kind(); got != "fake" {
		t.Fatalf("Kind() = %q, want %q", got, "fake")
	}
}
