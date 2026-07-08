package closer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/horizoonn/factory-platform/platform/pkg/logger"
)

const shutdownTimeout = 5 * time.Second

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

type Closer struct {
	mu     sync.Mutex
	once   sync.Once
	done   chan struct{}
	funcs  []func(context.Context) error
	logger Logger
}

var globalCloser = NewWithLogger(logger.NewNop())

func AddNamed(name string, f func(context.Context) error) {
	globalCloser.AddNamed(name, f)
}

func Add(f ...func(context.Context) error) {
	globalCloser.Add(f...)
}

func CloseAll(ctx context.Context) error {
	return globalCloser.CloseAll(ctx)
}

func SetLogger(l Logger) {
	globalCloser.SetLogger(l)
}

func Configure(signals ...os.Signal) {
	go globalCloser.handleSignals(signals...)
}

func New(signals ...os.Signal) *Closer {
	return NewWithLogger(logger.Default(), signals...)
}

func NewWithLogger(logger Logger, signals ...os.Signal) *Closer {
	c := &Closer{
		done:   make(chan struct{}),
		logger: logger,
	}

	if len(signals) > 0 {
		go c.handleSignals(signals...)
	}

	return c
}

func (c *Closer) SetLogger(l Logger) {
	c.logger = l
}

func (c *Closer) handleSignals(signals ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signals...)
	defer signal.Stop(ch)

	select {
	case <-ch:
		c.logger.Info(context.Background(), "shutdown signal received")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		if err := c.CloseAll(shutdownCtx); err != nil {
			c.logger.Error(context.Background(), "shutdown failed", zap.Error(err))
		}

	case <-c.done:
	}
}

func (c *Closer) AddNamed(name string, f func(context.Context) error) {
	c.Add(func(ctx context.Context) error {
		start := time.Now()
		c.logger.Info(ctx, "closing resource", zap.String("resource", name))

		err := f(ctx)
		duration := time.Since(start)
		if err != nil {
			c.logger.Error(
				ctx,
				"close resource failed",
				zap.String("resource", name),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		} else {
			c.logger.Info(
				ctx,
				"resource closed",
				zap.String("resource", name),
				zap.Duration("duration", duration),
			)
		}
		return err
	})
}

func (c *Closer) Add(f ...func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, f...)
}

func (c *Closer) CloseAll(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			c.logger.Info(ctx, "no shutdown functions registered")
			return
		}

		c.logger.Info(ctx, "shutdown started", zap.Int("functions", len(funcs)))

		var group errgroup.Group

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]
			group.Go(func() (err error) {
				defer func() {
					if r := recover(); r != nil {
						c.logger.Error(ctx, "shutdown function panicked", zap.Any("panic", r))
						err = fmt.Errorf("panic recovered in closer: %v", r)
					}
				}()

				if err := f(ctx); err != nil {
					c.logger.Error(ctx, "shutdown function failed", zap.Error(err))
					return err
				}

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			result = err
			return
		}

		if err := ctx.Err(); err != nil {
			c.logger.Warn(ctx, "shutdown context canceled", zap.Error(err))
			result = err
			return
		}

		c.logger.Info(ctx, "shutdown completed")
	})

	return result
}
