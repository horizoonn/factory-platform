//go:build e2e

package e2e

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func TestInventoryService_GetPart(t *testing.T) {
	env := requireTestEnv(t)
	env.ClearParts(t)

	expected := env.InsertPart(t, basePartFixture())
	client, closeClient := newInventoryClient(t, env)
	defer closeClient()

	resp, err := client.GetPart(testContext(t), &inventorypb.GetPartRequest{
		Uuid: expected.GetUuid(),
	})

	require.NoError(t, err)
	require.NotNil(t, resp.GetPart())
	assert.Equal(t, expected.GetUuid(), resp.GetPart().GetUuid())
	assert.Equal(t, expected.GetName(), resp.GetPart().GetName())
	assert.Equal(t, expected.GetDescription(), resp.GetPart().GetDescription())
	assert.Equal(t, expected.GetPrice(), resp.GetPart().GetPrice())
	assert.Equal(t, expected.GetStockQuantity(), resp.GetPart().GetStockQuantity())
	assert.Equal(t, expected.GetCategory(), resp.GetPart().GetCategory())
	assert.Equal(t, expected.GetDimensions(), resp.GetPart().GetDimensions())
	assert.Equal(t, expected.GetManufacturer(), resp.GetPart().GetManufacturer())
	assert.ElementsMatch(t, expected.GetTags(), resp.GetPart().GetTags())
	assert.Equal(t, expected.GetMetadata()["serial"].GetStringValue(), resp.GetPart().GetMetadata()["serial"].GetStringValue())
	assert.NotNil(t, resp.GetPart().GetCreatedAt())
	assert.NotNil(t, resp.GetPart().GetUpdatedAt())
}

func TestInventoryService_GetPartNotFound(t *testing.T) {
	env := requireTestEnv(t)
	env.ClearParts(t)

	client, closeClient := newInventoryClient(t, env)
	defer closeClient()

	_, err := client.GetPart(testContext(t), &inventorypb.GetPartRequest{
		Uuid: uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982404").String(),
	})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestInventoryService_ListParts(t *testing.T) {
	env := requireTestEnv(t)
	env.ClearParts(t)

	engineFixture := basePartFixture()
	engine := env.InsertPart(t, engineFixture)
	fuelFixture := basePartFixture()
	fuelFixture.ID = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982002")
	fuelFixture.Name = "fuel cell"
	fuelFixture.Category = inventorypb.Category_CATEGORY_FUEL
	fuelFixture.Manufacturer = &inventorypb.Manufacturer{
		Name:    "Fuel Corp",
		Country: "Germany",
		Website: "https://fuel.example.test",
	}
	fuelFixture.Tags = []string{"fuel"}
	env.InsertPart(t, fuelFixture)

	client, closeClient := newInventoryClient(t, env)
	defer closeClient()

	resp, err := client.ListParts(testContext(t), &inventorypb.ListPartsRequest{
		Filter: &inventorypb.PartsFilter{
			Categories:            []inventorypb.Category{inventorypb.Category_CATEGORY_ENGINE},
			ManufacturerCountries: []string{"USA"},
			Tags:                  []string{"critical"},
		},
	})

	require.NoError(t, err)
	require.Len(t, resp.GetParts(), 1)
	assert.Equal(t, engine.GetUuid(), resp.GetParts()[0].GetUuid())
	assert.Equal(t, engine.GetName(), resp.GetParts()[0].GetName())
}

func TestInventoryHTTP_ListParts(t *testing.T) {
	env := requireTestEnv(t)
	env.ClearParts(t)

	engine := env.InsertPart(t, basePartFixture())
	httpAddress, err := env.App.AddressFor(inventoryHTTPPort)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(
		testContext(t),
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/parts?filter.categories=CATEGORY_ENGINE&filter.tags=critical", httpAddress),
		http.NoBody,
	)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))

	var result inventorypb.ListPartsResponse
	require.NoError(t, protojson.Unmarshal(body, &result))
	require.Len(t, result.GetParts(), 1)
	assert.Equal(t, engine.GetUuid(), result.GetParts()[0].GetUuid())
	assert.Equal(t, engine.GetName(), result.GetParts()[0].GetName())
}

func newInventoryClient(t *testing.T, env *TestEnvironment) (inventorypb.InventoryServiceClient, func()) {
	t.Helper()

	conn, err := grpc.NewClient(
		env.App.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return inventorypb.NewInventoryServiceClient(conn), func() {
		require.NoError(t, conn.Close())
	}
}
