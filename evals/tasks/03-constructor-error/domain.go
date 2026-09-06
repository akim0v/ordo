package main

import "errors"

// ErrFlaky is returned by NewFlaky when the config asks it to fail.
var ErrFlaky = errors.New("flaky construction failed")

type Config struct {
	Fail bool
}

type Flaky struct{}

// NewFlaky is a constructor of the (T, error) shape.
func NewFlaky(cfg *Config) (*Flaky, error) {
	if cfg.Fail {
		return nil, ErrFlaky
	}
	return &Flaky{}, nil
}

func main() {}
