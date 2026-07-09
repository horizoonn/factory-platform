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
	if env.OrderApp != nil {
		err = errors.Join(err, env.OrderApp.Terminate(ctx))
	}
	if env.OrderPool != nil {
		env.OrderPool.Close()
	}
	if env.OrderPostgres != nil {
		err = errors.Join(err, env.OrderPostgres.Terminate(ctx))
	}
	if env.PaymentApp != nil {
		err = errors.Join(err, env.PaymentApp.Terminate(ctx))
	}
	if env.InventoryApp != nil {
		err = errors.Join(err, env.InventoryApp.Terminate(ctx))
	}
	if env.InventoryPool != nil {
		env.InventoryPool.Close()
	}
	if env.InventoryPostgres != nil {
		err = errors.Join(err, env.InventoryPostgres.Terminate(ctx))
	}
	if env.Network != nil {
		err = errors.Join(err, env.Network.Remove(ctx))
	}

	return err
}
