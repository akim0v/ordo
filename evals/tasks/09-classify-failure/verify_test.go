package main

import (
	"errors"
	"testing"
)

func TestClassification(t *testing.T) {
	err := BadContainer()
	if err == nil {
		t.Fatal("BadContainer() must return a registration error")
	}

	if !IsNotAFunction(err) {
		t.Fatalf("IsNotAFunction(%v) = false, want true", err)
	}

	idx, ok := FaultIndex(err)
	if !ok {
		t.Fatalf("FaultIndex(%v) reported no fault", err)
	}
	if idx != 0 {
		t.Fatalf("FaultIndex = %d, want 0", idx)
	}
}

func TestClassificationRejectsUnrelatedErrors(t *testing.T) {
	other := errors.New("unrelated")

	if IsNotAFunction(other) {
		t.Fatal("IsNotAFunction(unrelated) = true, want false")
	}
	if _, ok := FaultIndex(other); ok {
		t.Fatal("FaultIndex(unrelated) reported a fault")
	}
}
