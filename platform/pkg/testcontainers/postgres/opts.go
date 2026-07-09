package postgres

import "github.com/testcontainers/testcontainers-go"

type Option func(*Config)

func WithImage(image string) Option {
	return func(c *Config) {
		c.Image = image
	}
}

func WithDatabase(database string) Option {
	return func(c *Config) {
		c.Database = database
	}
}

func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithContainerCustomizers(customizers ...testcontainers.ContainerCustomizer) Option {
	return func(c *Config) {
		c.ContainerCustomizers = append(c.ContainerCustomizers, customizers...)
	}
}
