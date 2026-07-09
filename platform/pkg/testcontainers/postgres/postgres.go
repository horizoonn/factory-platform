package postgres

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"

	pgxpool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool/pgx"
)

type Container struct {
	container *tcpostgres.PostgresContainer
	cfg       Config
}

func NewContainer(ctx context.Context, opts ...Option) (*Container, error) {
	cfg := buildConfig(opts...)

	containerOpts := []testcontainers.ContainerCustomizer{
		tcpostgres.WithDatabase(cfg.Database),
		tcpostgres.WithUsername(cfg.Username),
		tcpostgres.WithPassword(cfg.Password),
		tcpostgres.BasicWaitStrategies(),
	}
	containerOpts = append(containerOpts, cfg.ContainerCustomizers...)

	container, err := tcpostgres.Run(
		ctx,
		cfg.Image,
		containerOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	success := false
	defer func() {
		if !success {
			if err := testcontainers.TerminateContainer(container); err != nil {
				cfg.Logger.Error(ctx, "failed to terminate postgres container", zap.Error(err))
			}
		}
	}()

	if _, err := container.ConnectionString(ctx, "sslmode=disable"); err != nil {
		return nil, fmt.Errorf("get postgres connection string: %w", err)
	}

	cfg.Logger.Info(ctx, "postgres container started", zap.String("image", cfg.Image))
	success = true

	return &Container{
		container: container,
		cfg:       cfg,
	}, nil
}

func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	connString, err := c.container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", fmt.Errorf("get postgres connection string: %w", err)
	}

	return connString, nil
}

func (c *Container) PgxPoolConfig(ctx context.Context) (pgxpool.Config, error) {
	connString, err := c.ConnectionString(ctx)
	if err != nil {
		return pgxpool.Config{}, err
	}

	parsed, err := url.Parse(connString)
	if err != nil {
		return pgxpool.Config{}, fmt.Errorf("parse postgres connection string: %w", err)
	}

	password, _ := parsed.User.Password()

	return pgxpool.Config{
		Host:            parsed.Hostname(),
		Port:            parsed.Port(),
		User:            parsed.User.Username(),
		Password:        password,
		Database:        strings.TrimPrefix(parsed.Path, "/"),
		SSLMode:         parsed.Query().Get("sslmode"),
		Timeout:         DefaultOperationTimeout,
		MaxConns:        DefaultMaxConns,
		MinConns:        DefaultMinConns,
		MaxConnIdleTime: DefaultMaxConnIdleTime,
	}, nil
}

func (c *Container) Terminate(ctx context.Context) error {
	if err := testcontainers.TerminateContainer(c.container); err != nil {
		return fmt.Errorf("terminate postgres container: %w", err)
	}

	c.cfg.Logger.Info(ctx, "postgres container terminated")

	return nil
}
