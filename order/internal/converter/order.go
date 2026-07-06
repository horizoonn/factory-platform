package converter

import (
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	orderopenapi "github.com/horizoonn/factory-platform/shared/pkg/openapi/order/v1"
)

func OrderToOpenAPI(order domain.Order) (*orderopenapi.OrderDto, error) {
	status, err := OrderStatusToOpenAPI(order.Status)
	if err != nil {
		return nil, err
	}

	dto := &orderopenapi.OrderDto{
		OrderUUID:  order.ID,
		UserUUID:   order.UserID,
		PartUuids:  order.PartIDs,
		TotalPrice: order.TotalPrice,
		Status:     status,
	}

	if order.TransactionID != nil {
		dto.TransactionUUID = orderopenapi.NewOptUUID(*order.TransactionID)
	}

	if order.PaymentMethod != nil {
		pm, err := PaymentMethodToOpenAPI(*order.PaymentMethod)
		if err != nil {
			return nil, err
		}
		dto.PaymentMethod = orderopenapi.NewOptPaymentMethod(pm)
	}

	return dto, nil
}

func OrderStatusToOpenAPI(status domain.OrderStatus) (orderopenapi.OrderStatus, error) {
	switch status {
	case domain.OrderStatusPendingPayment:
		return orderopenapi.OrderStatusPENDINGPAYMENT, nil
	case domain.OrderStatusPaid:
		return orderopenapi.OrderStatusPAID, nil
	case domain.OrderStatusCancelled:
		return orderopenapi.OrderStatusCANCELLED, nil
	default:
		return "", fmt.Errorf("unknown order status: %q", status)
	}
}

func PaymentMethodToDomain(method orderopenapi.PaymentMethod) (domain.PaymentMethod, error) {
	switch method {
	case orderopenapi.PaymentMethodCARD:
		return domain.PaymentMethodCard, nil
	case orderopenapi.PaymentMethodSBP:
		return domain.PaymentMethodSBP, nil
	case orderopenapi.PaymentMethodCREDITCARD:
		return domain.PaymentMethodCreditCard, nil
	case orderopenapi.PaymentMethodINVESTORMONEY:
		return domain.PaymentMethodInvestorMoney, nil
	default:
		return domain.PaymentMethodUnknown, domain.ErrInvalidPaymentMethod
	}
}

func PaymentMethodToOpenAPI(method domain.PaymentMethod) (orderopenapi.PaymentMethod, error) {
	switch method {
	case domain.PaymentMethodCard:
		return orderopenapi.PaymentMethodCARD, nil
	case domain.PaymentMethodSBP:
		return orderopenapi.PaymentMethodSBP, nil
	case domain.PaymentMethodCreditCard:
		return orderopenapi.PaymentMethodCREDITCARD, nil
	case domain.PaymentMethodInvestorMoney:
		return orderopenapi.PaymentMethodINVESTORMONEY, nil
	default:
		return "", fmt.Errorf("unknown payment method: %q", method)
	}
}
