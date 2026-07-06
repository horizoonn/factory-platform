package paymentv1

import (
	"github.com/horizoonn/factory-platform/payment/internal/domain"
	paymentpb "github.com/horizoonn/factory-platform/shared/pkg/proto/payment/v1"
)

func paymentMethodToDomain(method paymentpb.PaymentMethod) (domain.PaymentMethod, error) {
	switch method {
	case paymentpb.PaymentMethod_PAYMENT_METHOD_CARD:
		return domain.PaymentMethodCard, nil
	case paymentpb.PaymentMethod_PAYMENT_METHOD_SBP:
		return domain.PaymentMethodSBP, nil
	case paymentpb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		return domain.PaymentMethodCreditCard, nil
	case paymentpb.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY:
		return domain.PaymentMethodInvestorMoney, nil
	default:
		return domain.PaymentMethodUnknown, domain.ErrInvalidPaymentMethod
	}
}
