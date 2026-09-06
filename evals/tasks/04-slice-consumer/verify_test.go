package main

import (
	"slices"
	"testing"
)

func TestDispatcherReceivesEveryNotifier(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	d, err := c.GetService[*Dispatcher]()
	if err != nil {
		t.Fatalf("resolve *Dispatcher: %v", err)
	}

	want := []string{"email", "sms"}
	if got := d.Names(); !slices.Equal(got, want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
}
