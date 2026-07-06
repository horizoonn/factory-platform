package paymentv1

import (
	"context"

	servicedto "github.com/horizoonn/factory-platform/payment/internal/service/dto"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

type Server struct {
	paymentpb.UnimplementedPaymentServiceServer

	service PaymentService
}

type PaymentService interface {
	PayOrder(ctx context.Context, req servicedto.PayOrderRequest) (servicedto.PayOrderResponse, error)
}

func NewServer(service PaymentService) *Server {
	return &Server{
		service: service,
	}
}
