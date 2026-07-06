package paymentv1

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/client/converter"
	"github.com/horizoonn/factory-platform/order/internal/client/dto"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

func (c *Client) payOrder(ctx context.Context, req dto.PayOrderRequest) (dto.PayOrderResponse, error) {
	resp, err := c.client.PayOrder(ctx, &paymentpb.PayOrderRequest{
		OrderUuid:     req.OrderID.String(),
		UserUuid:      req.UserID.String(),
		PaymentMethod: converter.PaymentMethodToProto(req.PaymentMethod),
	})
	if err != nil {
		return dto.PayOrderResponse{}, fmt.Errorf("pay order grpc call: %w", err)
	}

	transactionID, err := uuid.Parse(resp.GetTransactionUuid())
	if err != nil {
		return dto.PayOrderResponse{}, fmt.Errorf("parse transaction uuid: %w", err)
	}

	return dto.PayOrderResponse{
		TransactionID: transactionID,
	}, nil
}
