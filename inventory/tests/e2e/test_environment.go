//go:build e2e

package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

type partFixture struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Price         float64
	StockQuantity int64
	Category      inventorypb.Category
	Dimensions    *inventorypb.Dimensions
	Manufacturer  *inventorypb.Manufacturer
	Tags          []string
	Metadata      map[string]*inventorypb.Value
}

func (env *TestEnvironment) InsertPart(t *testing.T, fixture partFixture) *inventorypb.Part {
	t.Helper()

	ctx := testContext(t)
	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	updatedAt := createdAt
	metadata := fixture.Metadata
	if metadata == nil {
		metadata = map[string]*inventorypb.Value{}
	}
	tags := fixture.Tags
	if tags == nil {
		tags = []string{}
	}

	_, err := env.Pool.Exec(
		ctx, `
		INSERT INTO platform.parts (
			id, name, description, price, stock_quantity, category,
			dimensions, manufacturer, tags, metadata, created_at, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7::jsonb, $8::jsonb, $9, $10::jsonb, $11, $12
		)
		`,
		fixture.ID,
		fixture.Name,
		fixture.Description,
		fixture.Price,
		fixture.StockQuantity,
		int32(fixture.Category),
		marshalDimensions(t, fixture.Dimensions),
		marshalManufacturer(t, fixture.Manufacturer),
		tags,
		marshalMetadata(t, metadata),
		createdAt,
		updatedAt,
	)
	require.NoError(t, err)

	return &inventorypb.Part{
		Uuid:          fixture.ID.String(),
		Name:          fixture.Name,
		Description:   fixture.Description,
		Price:         fixture.Price,
		StockQuantity: fixture.StockQuantity,
		Category:      fixture.Category,
		Dimensions:    fixture.Dimensions,
		Manufacturer:  fixture.Manufacturer,
		Tags:          tags,
		Metadata:      metadata,
	}
}

func (env *TestEnvironment) ClearParts(t *testing.T) {
	t.Helper()

	_, err := env.Pool.Exec(testContext(t), "TRUNCATE TABLE platform.parts CASCADE")
	require.NoError(t, err)
}

func basePartFixture() partFixture {
	return partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982001"),
		Name:          "warp engine",
		Description:   "primary engine",
		Price:         1250.50,
		StockQuantity: 7,
		Category:      inventorypb.Category_CATEGORY_ENGINE,
		Dimensions: &inventorypb.Dimensions{
			Length: 12.5,
			Width:  3.2,
			Height: 4.8,
			Weight: 9.1,
		},
		Manufacturer: &inventorypb.Manufacturer{
			Name:    "ACME",
			Country: "USA",
			Website: "https://example.test",
		},
		Tags: []string{"engine", "critical"},
		Metadata: map[string]*inventorypb.Value{
			"serial": {
				Kind: &inventorypb.Value_StringValue{StringValue: "SN-001"},
			},
			"revision": {
				Kind: &inventorypb.Value_Int64Value{Int64Value: 3},
			},
			"certified": {
				Kind: &inventorypb.Value_BoolValue{BoolValue: true},
			},
		},
	}
}

func marshalDimensions(t *testing.T, dimensions *inventorypb.Dimensions) []byte {
	t.Helper()
	if dimensions == nil {
		return nil
	}

	return marshalJSON(t, struct {
		Length float64 `json:"length"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
		Weight float64 `json:"weight"`
	}{
		Length: dimensions.GetLength(),
		Width:  dimensions.GetWidth(),
		Height: dimensions.GetHeight(),
		Weight: dimensions.GetWeight(),
	})
}

func marshalManufacturer(t *testing.T, manufacturer *inventorypb.Manufacturer) []byte {
	t.Helper()
	if manufacturer == nil {
		return nil
	}

	return marshalJSON(t, struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Website string `json:"website"`
	}{
		Name:    manufacturer.GetName(),
		Country: manufacturer.GetCountry(),
		Website: manufacturer.GetWebsite(),
	})
}

func marshalMetadata(t *testing.T, metadata map[string]*inventorypb.Value) []byte {
	t.Helper()

	values := make(map[string]jsonValue, len(metadata))
	for key, value := range metadata {
		values[key] = protoValueToJSON(value)
	}

	return marshalJSON(t, values)
}

func protoValueToJSON(value *inventorypb.Value) jsonValue {
	if value == nil {
		return jsonValue{}
	}

	switch typed := value.GetKind().(type) {
	case *inventorypb.Value_StringValue:
		return jsonValue{String: ptr(typed.StringValue)}
	case *inventorypb.Value_Int64Value:
		return jsonValue{Int64: ptr(typed.Int64Value)}
	case *inventorypb.Value_DoubleValue:
		return jsonValue{Float64: ptr(typed.DoubleValue)}
	case *inventorypb.Value_BoolValue:
		return jsonValue{Bool: ptr(typed.BoolValue)}
	default:
		return jsonValue{}
	}
}

type jsonValue struct {
	String  *string  `json:"string,omitempty"`
	Int64   *int64   `json:"int64,omitempty"`
	Float64 *float64 `json:"float64,omitempty"`
	Bool    *bool    `json:"bool,omitempty"`
}

func marshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	require.NoError(t, err)

	return data
}

func ptr[T any](value T) *T {
	return &value
}
