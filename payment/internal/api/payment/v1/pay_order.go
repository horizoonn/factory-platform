package paymentv1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/horizoonn/factory-platform/payment/internal/converter"
	"github.com/horizoonn/factory-platform/payment/internal/domain"
	servicedto "github.com/horizoonn/factory-platform/payment/internal/service/dto"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

func (s *Server) PayOrder(ctx context.Context, req *paymentpb.PayOrderRequest) (*paymentpb.PayOrderResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order uuid")
	}

	userID, err := uuid.Parse(req.GetUserUuid())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user uuid")
	}

	paymentMethod, err := converter.PaymentMethodToDomain(req.GetPaymentMethod())
	if err != nil {
		if req.GetPaymentMethod() == paymentpb.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED {
			return nil, status.Error(codes.InvalidArgument, "payment_method is required")
		}

		return nil, status.Error(codes.InvalidArgument, "invalid payment_method")
	}

	result, err := s.service.PayOrder(ctx, servicedto.PayOrderRequest{
		OrderID:       orderID,
		UserID:        userID,
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPaymentMethod) ||
			errors.Is(err, domain.ErrUserIDRequired) ||
			errors.Is(err, domain.ErrOrderIDRequired) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.Canceled, err.Error())
		}

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, err.Error())
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &paymentpb.PayOrderResponse{
		TransactionUuid: result.TransactionID.String(),
	}, nil
}
