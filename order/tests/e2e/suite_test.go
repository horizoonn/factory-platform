//go:build e2e

package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var testEnv *TestEnvironment

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), setupTimeout)
	defer cancel()

	env, err := setupTestEnvironment(ctx)
	if err != nil {
		panic(err)
	}
	testEnv = env

	code := m.Run()

	teardownCtx, teardownCancel := context.WithTimeout(context.Background(), teardownTimeout)
	defer teardownCancel()

	if err := teardownTestEnvironment(teardownCtx, env); err != nil {
		panic(err)
	}

	os.Exit(code)
}

func testContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	t.Cleanup(cancel)

	return ctx
}

func requireTestEnv(t *testing.T) *TestEnvironment {
	t.Helper()

	require.NotNil(t, testEnv)

	return testEnv
}
