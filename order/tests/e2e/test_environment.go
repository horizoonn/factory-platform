//go:build e2e

package e2e

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

type inventoryPartFixture struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Price         float64
	StockQuantity int64
	Category      inventorypb.Category
	Tags          []string
}

func (env *TestEnvironment) ClearData(t *testing.T) {
	t.Helper()

	_, err := env.OrderPool.Exec(testContext(t), "TRUNCATE TABLE platform.orders CASCADE")
	require.NoError(t, err)

	_, err = env.InventoryPool.Exec(testContext(t), "TRUNCATE TABLE platform.parts CASCADE")
	require.NoError(t, err)
}

func (env *TestEnvironment) InsertInventoryPart(t *testing.T, fixture inventoryPartFixture) inventoryPartFixture {
	t.Helper()

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	updatedAt := createdAt
	tags := fixture.Tags
	if tags == nil {
		tags = []string{}
	}

	_, err := env.InventoryPool.Exec(testContext(t), `
		INSERT INTO platform.parts (
			id, name, description, price, stock_quantity, category,
			dimensions, manufacturer, tags, metadata, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			NULL, NULL, $7, '{}'::jsonb, $8, $9
		)
		`,
		fixture.ID,
		fixture.Name,
		fixture.Description,
		fixture.Price,
		fixture.StockQuantity,
		int32(fixture.Category),
		tags,
		createdAt,
		updatedAt,
	)
	require.NoError(t, err)

	fixture.Tags = tags

	return fixture
}

func baseInventoryPart() inventoryPartFixture {
	return inventoryPartFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982201"),
		Name:          "warp engine",
		Description:   "primary engine",
		Price:         1250.50,
		StockQuantity: 7,
		Category:      inventorypb.Category_CATEGORY_ENGINE,
		Tags:          []string{"engine", "critical"},
	}
}
