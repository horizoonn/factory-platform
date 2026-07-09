//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	partrepo "github.com/horizoonn/factory-platform/inventory/internal/repository/part"
)

func TestPartRepository_GetPart(t *testing.T) {
	testEnv.truncateParts(t)

	repository := partrepo.NewRepository(testEnv.pool)
	expected := insertPart(t, partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981001"),
		Name:          "warp engine",
		Description:   "primary engine",
		Price:         1250.50,
		StockQuantity: 7,
		Category:      domain.CATEGORY_ENGINE,
		Dimensions: &domain.Dimensions{
			Length: 12.5,
			Width:  3.2,
			Height: 4.8,
			Weight: 9.1,
		},
		Manufacturer: &domain.Manufacturer{
			Name:    "ACME",
			Country: "USA",
			Website: "https://example.test",
		},
		Tags: []string{"engine", "critical"},
		Metadata: map[string]*domain.Value{
			"serial":    {String: ptr("SN-001")},
			"revision":  {Int64: ptr(int64(3))},
			"certified": {Bool: ptr(true)},
		},
	})

	actual, err := repository.GetPart(testContext(t), expected.ID)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description)
	assert.Equal(t, expected.Price, actual.Price)
	assert.Equal(t, expected.StockQuantity, actual.StockQuantity)
	assert.Equal(t, expected.Category, actual.Category)
	assert.Equal(t, expected.Dimensions, actual.Dimensions)
	assert.Equal(t, expected.Manufacturer, actual.Manufacturer)
	assert.ElementsMatch(t, expected.Tags, actual.Tags)
	require.Contains(t, actual.Metadata, "serial")
	assert.Equal(t, "SN-001", *actual.Metadata["serial"].String)
	require.Contains(t, actual.Metadata, "revision")
	assert.Equal(t, int64(3), *actual.Metadata["revision"].Int64)
	require.Contains(t, actual.Metadata, "certified")
	assert.Equal(t, true, *actual.Metadata["certified"].Bool)
}

func TestPartRepository_GetPartNullableJSONAndEmptyCollections(t *testing.T) {
	testEnv.truncateParts(t)

	repository := partrepo.NewRepository(testEnv.pool)
	expected := insertPart(t, partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981002"),
		Name:          `seal "alpha"`,
		Description:   "part without optional json fields",
		Price:         0.01,
		StockQuantity: 0,
		Category:      domain.CATEGORY_PORTHOLE,
		Tags:          []string{},
		Metadata:      map[string]*domain.Value{},
	})

	actual, err := repository.GetPart(testContext(t), expected.ID)

	require.NoError(t, err)
	assert.Nil(t, actual.Dimensions)
	assert.Nil(t, actual.Manufacturer)
	assert.Empty(t, actual.Tags)
	assert.Empty(t, actual.Metadata)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Price, actual.Price)
}

func TestPartRepository_GetPartNotFound(t *testing.T) {
	testEnv.truncateParts(t)

	repository := partrepo.NewRepository(testEnv.pool)

	_, err := repository.GetPart(testContext(t), uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981404"))

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrNotFound))
}

func TestPartRepository_ListParts(t *testing.T) {
	testEnv.truncateParts(t)

	repository := partrepo.NewRepository(testEnv.pool)
	engine := insertPart(t, partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981101"),
		Name:          "warp engine",
		Description:   "primary engine",
		Price:         1250.50,
		StockQuantity: 7,
		Category:      domain.CATEGORY_ENGINE,
		Manufacturer: &domain.Manufacturer{
			Name:    "ACME",
			Country: "USA",
			Website: "https://example.test",
		},
		Tags: []string{"engine", "critical"},
	})
	insertPart(t, partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981102"),
		Name:          "fuel cell",
		Description:   "backup fuel",
		Price:         300,
		StockQuantity: 11,
		Category:      domain.CATEGORY_FUEL,
		Manufacturer: &domain.Manufacturer{
			Name:    "Fuel Corp",
			Country: "Germany",
			Website: "https://fuel.example.test",
		},
		Tags: []string{"fuel"},
	})

	parts, err := repository.ListParts(testContext(t), domain.PartsFilter{
		Categories:            []domain.Category{domain.CATEGORY_ENGINE},
		ManufacturerCountries: []string{"USA"},
		Tags:                  []string{"critical"},
	})

	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, engine.ID, parts[0].ID)
	assert.Equal(t, engine.Name, parts[0].Name)
}

func TestPartRepository_ListPartsByName(t *testing.T) {
	testEnv.truncateParts(t)

	repository := partrepo.NewRepository(testEnv.pool)
	expected := insertPart(t, partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981201"),
		Name:          `seal "alpha" / beta`,
		Description:   "special characters in name",
		Price:         42,
		StockQuantity: 3,
		Category:      domain.CATEGORY_PORTHOLE,
		Tags:          []string{"seal"},
	})
	insertPart(t, partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981202"),
		Name:          "seal gamma",
		Description:   "another seal",
		Price:         24,
		StockQuantity: 4,
		Category:      domain.CATEGORY_PORTHOLE,
		Tags:          []string{"seal"},
	})

	parts, err := repository.ListParts(testContext(t), domain.PartsFilter{
		Names: []string{expected.Name},
	})

	require.NoError(t, err)
	require.Len(t, parts, 1)
	assert.Equal(t, expected.ID, parts[0].ID)
	assert.Equal(t, expected.Name, parts[0].Name)
}

func TestPartRepository_ListPartsEmptyResult(t *testing.T) {
	testEnv.truncateParts(t)

	repository := partrepo.NewRepository(testEnv.pool)
	insertPart(t, partFixture{
		ID:            uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f981301"),
		Name:          "warp engine",
		Description:   "primary engine",
		Price:         1250.50,
		StockQuantity: 7,
		Category:      domain.CATEGORY_ENGINE,
		Tags:          []string{"engine"},
	})

	parts, err := repository.ListParts(testContext(t), domain.PartsFilter{
		Names: []string{"missing part"},
	})

	require.NoError(t, err)
	assert.Empty(t, parts)
}

type partFixture struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Price         float64
	StockQuantity int64
	Category      domain.Category
	Dimensions    *domain.Dimensions
	Manufacturer  *domain.Manufacturer
	Tags          []string
	Metadata      map[string]*domain.Value
}

func insertPart(t *testing.T, fixture partFixture) domain.Part {
	t.Helper()

	ctx := testContext(t)
	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	updatedAt := createdAt

	dimensionsJSON := marshalDimensions(t, fixture.Dimensions)
	manufacturerJSON := marshalManufacturer(t, fixture.Manufacturer)
	metadata := fixture.Metadata
	if metadata == nil {
		metadata = map[string]*domain.Value{}
	}
	metadataJSON := marshalMetadata(t, metadata)
	tags := fixture.Tags
	if tags == nil {
		tags = []string{}
	}

	_, err := testEnv.pool.Exec(ctx, `
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
		dimensionsJSON,
		manufacturerJSON,
		tags,
		metadataJSON,
		createdAt,
		updatedAt,
	)
	require.NoError(t, err)

	return domain.Part{
		ID:            fixture.ID,
		Name:          fixture.Name,
		Description:   fixture.Description,
		Price:         fixture.Price,
		StockQuantity: fixture.StockQuantity,
		Category:      fixture.Category,
		Dimensions:    fixture.Dimensions,
		Manufacturer:  fixture.Manufacturer,
		Tags:          tags,
		Metadata:      metadata,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
	}
}

func marshalDimensions(t *testing.T, dimensions *domain.Dimensions) []byte {
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
		Length: dimensions.Length,
		Width:  dimensions.Width,
		Height: dimensions.Height,
		Weight: dimensions.Weight,
	})
}

func marshalManufacturer(t *testing.T, manufacturer *domain.Manufacturer) []byte {
	t.Helper()
	if manufacturer == nil {
		return nil
	}

	return marshalJSON(t, struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Website string `json:"website"`
	}{
		Name:    manufacturer.Name,
		Country: manufacturer.Country,
		Website: manufacturer.Website,
	})
}

func marshalMetadata(t *testing.T, metadata map[string]*domain.Value) []byte {
	t.Helper()

	values := make(map[string]jsonValue, len(metadata))
	for key, value := range metadata {
		if value == nil {
			values[key] = jsonValue{}
			continue
		}
		values[key] = jsonValue{
			String:  value.String,
			Int64:   value.Int64,
			Float64: value.Float64,
			Bool:    value.Bool,
		}
	}

	return marshalJSON(t, values)
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

func testContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	return ctx
}

func ptr[T any](value T) *T {
	return &value
}
