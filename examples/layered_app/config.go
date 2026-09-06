package main

// Config is built by the application, not by the container, and handed over as
// a ready value.
type Config struct {
	AppName string
	Debug   bool
}

func LoadConfig() *Config {
	return &Config{
		AppName: "layered-app",
		Debug:   true,
	}
}
