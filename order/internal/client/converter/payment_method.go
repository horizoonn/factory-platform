package converter

import (
	"github.com/horizoonn/factory-platform/order/internal/domain"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

func PaymentMethodToProto(method domain.PaymentMethod) paymentpb.PaymentMethod {
	switch method {
	case domain.PaymentMethodCard:
		return paymentpb.PaymentMethod_PAYMENT_METHOD_CARD
	case domain.PaymentMethodSBP:
		return paymentpb.PaymentMethod_PAYMENT_METHOD_SBP
	case domain.PaymentMethodCreditCard:
		return paymentpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case domain.PaymentMethodInvestorMoney:
		return paymentpb.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY
	default:
		return paymentpb.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}
