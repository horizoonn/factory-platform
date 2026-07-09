//go:build e2e

package e2e

import (
	"context"
	"errors"
)

func teardownTestEnvironment(ctx context.Context, env *TestEnvironment) error {
	if env == nil {
		return nil
	}

	var err error
	if env.App != nil {
		err = errors.Join(err, env.App.Terminate(ctx))
	}
	if env.Pool != nil {
		env.Pool.Close()
	}
	if env.Postgres != nil {
		err = errors.Join(err, env.Postgres.Terminate(ctx))
	}
	if env.Network != nil {
		err = errors.Join(err, env.Network.Remove(ctx))
	}

	return err
}
