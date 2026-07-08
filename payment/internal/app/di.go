package app

import (
	paymentapi "github.com/horizoonn/factory-platform/payment/internal/api/payment/v1"
	paymentservice "github.com/horizoonn/factory-platform/payment/internal/service/payment"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

type diContainer struct {
	paymentV1API paymentpb.PaymentServiceServer

	paymentService paymentapi.PaymentService
}

func newDIContainer() *diContainer {
	return &diContainer{}
}

func (d *diContainer) PaymentV1API() paymentpb.PaymentServiceServer {
	if d.paymentV1API == nil {
		d.paymentV1API = paymentapi.NewServer(d.PaymentService())
	}

	return d.paymentV1API
}

func (d *diContainer) PaymentService() paymentapi.PaymentService {
	if d.paymentService == nil {
		d.paymentService = paymentservice.NewService()
	}

	return d.paymentService
}
