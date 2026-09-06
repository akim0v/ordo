package main

import "testing"

func TestKeyedResolution(t *testing.T) {
	c, err := BuildContainer()
	if err != nil {
		t.Fatalf("BuildContainer: %v", err)
	}

	for _, key := range []string{"redis", "memory"} {
		got, err := Get(c, key)
		if err != nil {
			t.Fatalf("Get(%q): %v", key, err)
		}
		if got.Name() != key {
			t.Fatalf("Get(%q).Name() = %q, want %q", key, got.Name(), key)
		}
	}
}
