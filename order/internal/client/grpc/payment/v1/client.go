package paymentv1

import (
	"context"

	"google.golang.org/grpc"

	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
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

func (c *Client) PayOrder(ctx context.Context, req servicedto.PaymentRequest) (servicedto.PaymentResponse, error) {
	return c.payOrder(ctx, req)
}
