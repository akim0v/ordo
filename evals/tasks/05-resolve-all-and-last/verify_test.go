package main

import (
	"slices"
	"testing"
)

func TestAllReturnsEveryRegistrationInOrder(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	all, err := All(c)
	if err != nil {
		t.Fatalf("All: %v", err)
	}

	names := make([]string, 0, len(all))
	for _, n := range all {
		names = append(names, n.Name())
	}

	want := []string{"email", "sms"}
	if !slices.Equal(names, want) {
		t.Fatalf("All() = %v, want %v", names, want)
	}
}

func TestLastReturnsTheFinalRegistration(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	last, err := Last(c)
	if err != nil {
		t.Fatalf("Last: %v", err)
	}

	if got := last.Name(); got != "sms" {
		t.Fatalf("Last().Name() = %q, want %q", got, "sms")
	}
}
