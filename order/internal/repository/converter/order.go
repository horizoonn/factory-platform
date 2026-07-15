package converter

import (
	"database/sql"
	"fmt"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	"github.com/horizoonn/factory-platform/order/internal/repository/model"
)

func OrderModelToDomain(m model.Order) (domain.Order, error) {
	order := domain.Order{
		ID:            m.ID,
		UserID:        m.UserID,
		PartIDs:       m.PartIDs,
		TotalPrice:    m.TotalPrice,
		TransactionID: m.TransactionID,
		Status:        domain.OrderStatus(m.Status),
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}

	if m.PaymentMethod.Valid {
		method := domain.PaymentMethod(m.PaymentMethod.String)
		if !method.Valid() {
			return domain.Order{}, fmt.Errorf(
				"invalid payment method %q: %w",
				m.PaymentMethod.String,
				domain.ErrInvalidPaymentMethod,
			)
		}
		order.PaymentMethod = &method
	}

	if !order.Status.Valid() {
		return domain.Order{}, fmt.Errorf(
			"invalid order status %q: %w",
			m.Status,
			domain.ErrInvalidOrderStatus,
		)
	}

	return order, nil
}

func DomainOrderToModel(order domain.Order) model.Order {
	m := model.Order{
		ID:         order.ID,
		UserID:     order.UserID,
		PartIDs:    order.PartIDs,
		TotalPrice: order.TotalPrice,
		Status:     string(order.Status),
	}

	if order.TransactionID != nil {
		m.TransactionID = order.TransactionID
	}

	if order.PaymentMethod != nil {
		m.PaymentMethod = sql.NullString{String: string(*order.PaymentMethod), Valid: true}
	}

	return m
}
