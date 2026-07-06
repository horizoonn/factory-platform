package paymentv1

import (
	"context"

	"google.golang.org/grpc"

	"github.com/horizoonn/factory-platform/order/internal/client/dto"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

type Client struct {
	client paymentpb.PaymentServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{
		client: paymentpb.NewPaymentServiceClient(conn),
	}
}

func (c *Client) PayOrder(ctx context.Context, req dto.PayOrderRequest) (dto.PayOrderResponse, error) {
	return c.payOrder(ctx, req)
}
