package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

const (
	defaultName           = "app"
	defaultPort           = "50051"
	defaultStartupTimeout = time.Minute
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

type Config struct {
	Name          string
	DockerfileDir string
	Dockerfile    string
	Port          string
	ExtraPorts    []string
	Env           map[string]string
	Networks      []string
	LogOutput     io.Writer
	StartupWait   wait.Strategy
	Logger        Logger
	KeepImage     bool
}

type Container struct {
	container     testcontainers.Container
	externalHost  string
	externalPorts map[string]string
	cfg           Config
}

func NewContainer(ctx context.Context, opts ...Option) (*Container, error) {
	cfg := Config{
		Name:          defaultName,
		Port:          defaultPort,
		Dockerfile:    "Dockerfile",
		DockerfileDir: ".",
		Env:           make(map[string]string),
		LogOutput:     io.Discard,
		StartupWait:   wait.ForListeningPort(defaultPort + "/tcp").WithStartupTimeout(defaultStartupTimeout),
		Logger:        logger.NewNop(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	exposedPorts := make([]string, 0, 1+len(cfg.ExtraPorts))
	exposedPorts = append(exposedPorts, cfg.Port+"/tcp")
	for _, port := range cfg.ExtraPorts {
		exposedPorts = append(exposedPorts, port+"/tcp")
	}

	req := testcontainers.ContainerRequest{
		Name: cfg.Name,
		FromDockerfile: testcontainers.FromDockerfile{
			Context:        cfg.DockerfileDir,
			Dockerfile:     cfg.Dockerfile,
			BuildLogWriter: cfg.LogOutput,
			KeepImage:      cfg.KeepImage,
		},
		Networks:           cfg.Networks,
		Env:                cfg.Env,
		ExposedPorts:       exposedPorts,
		WaitingFor:         cfg.StartupWait,
		HostConfigModifier: defaultHostConfig(),
	}

	genericContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("start app container: %w", err)
	}

	success := false
	defer func() {
		if !success {
			if err := testcontainers.TerminateContainer(genericContainer); err != nil {
				cfg.Logger.Error(ctx, "failed to terminate app container", zap.Error(err))
			}
		}
	}()

	externalPorts := make(map[string]string, len(exposedPorts))
	for _, port := range append([]string{cfg.Port}, cfg.ExtraPorts...) {
		mappedPort, err := genericContainer.MappedPort(ctx, port+"/tcp")
		if err != nil {
			return nil, fmt.Errorf("get mapped app port %s: %w", port, err)
		}
		externalPorts[port] = mappedPort.Port()
	}

	host, err := genericContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("get app host: %w", err)
	}

	go streamContainerLogs(ctx, genericContainer, cfg.LogOutput)

	cfg.Logger.Info(ctx, "app container started", zap.String("address", net.JoinHostPort(host, externalPorts[cfg.Port])))
	success = true

	return &Container{
		container:     genericContainer,
		externalHost:  host,
		externalPorts: externalPorts,
		cfg:           cfg,
	}, nil
}

func (c *Container) Address() string {
	return net.JoinHostPort(c.externalHost, c.externalPorts[c.cfg.Port])
}

func (c *Container) AddressFor(port string) (string, error) {
	externalPort, ok := c.externalPorts[port]
	if !ok {
		return "", fmt.Errorf("app port %s is not exposed", port)
	}

	return net.JoinHostPort(c.externalHost, externalPort), nil
}

func (c *Container) Terminate(ctx context.Context) error {
	if err := testcontainers.TerminateContainer(c.container); err != nil {
		return fmt.Errorf("terminate app container: %w", err)
	}

	return nil
}

func streamContainerLogs(ctx context.Context, container testcontainers.Container, out io.Writer) {
	logs, err := container.Logs(ctx)
	if err != nil {
		logger.Error(ctx, "get container logs", zap.Error(err))
		return
	}
	defer func() {
		if err := logs.Close(); err != nil {
			logger.Error(ctx, "close container logs", zap.Error(err))
		}
	}()

	if _, err := io.Copy(out, logs); err != nil && !errors.Is(err, io.EOF) {
		logger.Error(ctx, "copy container logs", zap.Error(err))
	}
}

func defaultHostConfig() func(*container.HostConfig) {
	return func(hc *container.HostConfig) {
		hc.AutoRemove = true
	}
}
