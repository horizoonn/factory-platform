package pgxpool

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `envconfig:"HOST" required:"true"`
	Port     string        `envconfig:"PORT" default:"5432"`
	User     string        `envconfig:"USER" required:"true"`
	Password string        `envconfig:"PASSWORD" required:"true"`
	Database string        `envconfig:"DB" required:"true"`
	SSLMode  string        `envconfig:"SSL_MODE" default:"disable"`
	Timeout  time.Duration `envconfig:"TIMEOUT" required:"true"`

	MaxConns        int32         `envconfig:"MAX_CONNS" default:"10"`
	MinConns        int32         `envconfig:"MIN_CONNS" default:"2"`
	MaxConnIdleTime time.Duration `envconfig:"MAX_CONN_IDLE_TIME" default:"5m"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("POSTGRES", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get Postgres connection pool config: %w", err)
		panic(err)
	}

	return config
}
