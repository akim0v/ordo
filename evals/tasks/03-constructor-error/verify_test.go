package main

import (
	"errors"
	"testing"
)

func TestConstructorSucceeds(t *testing.T) {
	c, err := BuildContainer(false)
	if err != nil {
		t.Fatalf("BuildContainer(false): %v", err)
	}

	if _, err := c.GetService[*Flaky](); err != nil {
		t.Fatalf("resolve *Flaky: %v", err)
	}
}

func TestConstructorErrorSurfacesOnResolution(t *testing.T) {
	c, err := BuildContainer(true)
	if err != nil {
		t.Fatalf("BuildContainer(true) must succeed; the graph is sound: %v", err)
	}

	if _, err := c.GetService[*Flaky](); !errors.Is(err, ErrFlaky) {
		t.Fatalf("resolve *Flaky: error = %v, want one wrapping ErrFlaky", err)
	}
}
