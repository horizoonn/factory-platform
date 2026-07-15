package app

import (
	"io"

	"github.com/testcontainers/testcontainers-go/wait"
)

type Option func(*Config)

func WithName(name string) Option {
	return func(c *Config) {
		c.Name = name
	}
}

func WithDockerfile(dir, file string) Option {
	return func(c *Config) {
		c.DockerfileDir = dir
		c.Dockerfile = file
	}
}

func WithPort(port string) Option {
	return func(c *Config) {
		c.Port = port
	}
}

func WithExtraPorts(ports ...string) Option {
	return func(c *Config) {
		c.ExtraPorts = append(c.ExtraPorts, ports...)
	}
}

func WithNetwork(name string) Option {
	return func(c *Config) {
		c.Networks = append(c.Networks, name)
	}
}

func WithEnv(env map[string]string) Option {
	return func(c *Config) {
		for key, value := range env {
			c.Env[key] = value
		}
	}
}

func WithLogOutput(out io.Writer) Option {
	return func(c *Config) {
		c.LogOutput = out
	}
}

func WithStartupWait(strategy wait.Strategy) Option {
	return func(c *Config) {
		c.StartupWait = strategy
	}
}

func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithKeepImage(keep bool) Option {
	return func(c *Config) {
		c.KeepImage = keep
	}
}
