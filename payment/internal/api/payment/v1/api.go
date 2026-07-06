package paymentv1

import (
	"context"

	paymentservice "github.com/horizoonn/factory-platform/payment/internal/service"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

type PaymentServer struct {
	paymentpb.UnimplementedPaymentServiceServer

	paymentService PaymentService
}

type PaymentService interface {
	PayOrder(ctx context.Context, req paymentservice.PayOrderRequest) (paymentservice.PayOrderResponse, error)
}

func NewPaymentServer(paymentService PaymentService) *PaymentServer {
	return &PaymentServer{
		paymentService: paymentService,
	}
}
