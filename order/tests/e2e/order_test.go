//go:build e2e

package e2e

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func TestOrderService_CreateGetPayOrder(t *testing.T) {
	env := requireTestEnv(t)
	env.ClearData(t)

	engine := env.InsertInventoryPart(t, baseInventoryPart())
	fuelFixture := baseInventoryPart()
	fuelFixture.ID = uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982202")
	fuelFixture.Name = "fuel cell"
	fuelFixture.Price = 300.25
	fuel := env.InsertInventoryPart(t, fuelFixture)

	client := newOrderClient(t, env)
	userID := uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982101")

	createRes, err := client.CreateOrder(testContext(t), &orderopenapi.CreateOrderRequest{
		UserUUID:  userID,
		PartUuids: []uuid.UUID{engine.ID, fuel.ID},
	})
	require.NoError(t, err)
	created := requireCreateOrderResponse(t, createRes)
	assert.NotEqual(t, uuid.Nil, created.OrderUUID)
	assert.InEpsilon(t, engine.Price+fuel.Price, created.TotalPrice, 0.0001)

	getRes, err := client.GetOrder(testContext(t), orderopenapi.GetOrderParams{
		OrderUUID: created.OrderUUID,
	})
	require.NoError(t, err)
	order := requireOrderDTO(t, getRes)
	assert.Equal(t, created.OrderUUID, order.OrderUUID)
	assert.Equal(t, userID, order.UserUUID)
	assert.ElementsMatch(t, []uuid.UUID{engine.ID, fuel.ID}, order.PartUuids)
	assert.InEpsilon(t, created.TotalPrice, order.TotalPrice, 0.0001)
	assert.Equal(t, orderopenapi.OrderStatusPENDINGPAYMENT, order.Status)
	assert.False(t, order.TransactionUUID.IsSet())
	assert.False(t, order.PaymentMethod.IsSet())

	payRes, err := client.PayOrder(testContext(t), &orderopenapi.PayOrderRequest{
		PaymentMethod: orderopenapi.PaymentMethodCARD,
	}, orderopenapi.PayOrderParams{
		OrderUUID: created.OrderUUID,
	})
	require.NoError(t, err)
	paid := requirePayOrderResponse(t, payRes)
	assert.NotEqual(t, uuid.Nil, paid.TransactionUUID)

	getPaidRes, err := client.GetOrder(testContext(t), orderopenapi.GetOrderParams{
		OrderUUID: created.OrderUUID,
	})
	require.NoError(t, err)
	paidOrder := requireOrderDTO(t, getPaidRes)
	assert.Equal(t, orderopenapi.OrderStatusPAID, paidOrder.Status)
	assert.Equal(t, paid.TransactionUUID, paidOrder.TransactionUUID.Value)
	assert.Equal(t, orderopenapi.PaymentMethodCARD, paidOrder.PaymentMethod.Value)
}

func TestOrderService_CancelOrder(t *testing.T) {
	env := requireTestEnv(t)
	env.ClearData(t)

	part := env.InsertInventoryPart(t, baseInventoryPart())
	client := newOrderClient(t, env)
	userID := uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982102")

	createRes, err := client.CreateOrder(testContext(t), &orderopenapi.CreateOrderRequest{
		UserUUID:  userID,
		PartUuids: []uuid.UUID{part.ID},
	})
	require.NoError(t, err)
	created := requireCreateOrderResponse(t, createRes)

	cancelRes, err := client.CancelOrder(testContext(t), orderopenapi.CancelOrderParams{
		OrderUUID: created.OrderUUID,
	})
	require.NoError(t, err)
	require.IsType(t, &orderopenapi.CancelOrderNoContent{}, cancelRes)

	getRes, err := client.GetOrder(testContext(t), orderopenapi.GetOrderParams{
		OrderUUID: created.OrderUUID,
	})
	require.NoError(t, err)
	cancelled := requireOrderDTO(t, getRes)
	assert.Equal(t, orderopenapi.OrderStatusCANCELLED, cancelled.Status)

	payRes, err := client.PayOrder(testContext(t), &orderopenapi.PayOrderRequest{
		PaymentMethod: orderopenapi.PaymentMethodCARD,
	}, orderopenapi.PayOrderParams{
		OrderUUID: created.OrderUUID,
	})
	require.NoError(t, err)
	require.IsType(t, &orderopenapi.ConflictError{}, payRes)
}

func TestOrderService_CreateOrderPartsNotFound(t *testing.T) {
	env := requireTestEnv(t)
	env.ClearData(t)

	client := newOrderClient(t, env)

	res, err := client.CreateOrder(testContext(t), &orderopenapi.CreateOrderRequest{
		UserUUID: uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982103"),
		PartUuids: []uuid.UUID{
			uuid.MustParse("7d4a1f4f-07cc-48b2-b7c7-f6201f982404"),
		},
	})

	require.NoError(t, err)
	require.IsType(t, &orderopenapi.BadRequestError{}, res)
}

func newOrderClient(t *testing.T, env *TestEnvironment) *orderopenapi.Client {
	t.Helper()

	client, err := orderopenapi.NewClient("http://" + env.OrderApp.Address())
	require.NoError(t, err)

	return client
}

func requireCreateOrderResponse(t *testing.T, res orderopenapi.CreateOrderRes) *orderopenapi.CreateOrderResponse {
	t.Helper()

	response, ok := res.(*orderopenapi.CreateOrderResponse)
	require.Truef(t, ok, "unexpected create order response type %T", res)

	return response
}

func requireOrderDTO(t *testing.T, res orderopenapi.GetOrderRes) *orderopenapi.OrderDto {
	t.Helper()

	response, ok := res.(*orderopenapi.OrderDto)
	require.Truef(t, ok, "unexpected get order response type %T", res)

	return response
}

func requirePayOrderResponse(t *testing.T, res orderopenapi.PayOrderRes) *orderopenapi.PayOrderResponse {
	t.Helper()

	response, ok := res.(*orderopenapi.PayOrderResponse)
	require.Truef(t, ok, "unexpected pay order response type %T", res)

	return response
}
