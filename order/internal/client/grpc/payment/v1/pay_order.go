package paymentv1

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/client/converter"
	servicedto "github.com/horizoonn/factory-platform/order/internal/service/dto"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

func (c *Client) payOrder(ctx context.Context, req servicedto.PaymentRequest) (servicedto.PaymentResponse, error) {
	resp, err := c.client.PayOrder(ctx, &paymentpb.PayOrderRequest{
		OrderUuid:     req.OrderID.String(),
		UserUuid:      req.UserID.String(),
		PaymentMethod: converter.PaymentMethodToProto(req.PaymentMethod),
	})
	if err != nil {
		return servicedto.PaymentResponse{}, fmt.Errorf("pay order grpc call: %w", err)
	}

	transactionID, err := uuid.Parse(resp.GetTransactionUuid())
	if err != nil {
		return servicedto.PaymentResponse{}, fmt.Errorf("parse transaction uuid: %w", err)
	}

	return servicedto.PaymentResponse{
		TransactionID: transactionID,
	}, nil
}
