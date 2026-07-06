package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	postgrespool "github.com/horizoonn/factory-platform/platform/pkg/database/postgres/pool"
)

type orderModel struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	PartIDs       []uuid.UUID
	TotalPrice    float64
	TransactionID *uuid.UUID
	PaymentMethod sql.NullString
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (m *orderModel) scan(row postgrespool.Row) error {
	return row.Scan(
		&m.ID,
		&m.UserID,
		&m.PartIDs,
		&m.TotalPrice,
		&m.TransactionID,
		&m.PaymentMethod,
		&m.Status,
		&m.CreatedAt,
		&m.UpdatedAt,
	)
}

func (m *orderModel) toDomain() (domain.Order, error) {
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
			return domain.Order{}, fmt.Errorf("invalid payment method %q: %w", m.PaymentMethod.String, domain.ErrInvalidPaymentMethod)
		}
		order.PaymentMethod = &method
	}

	if !order.Status.Valid() {
		return domain.Order{}, fmt.Errorf("invalid order status %q: %w", m.Status, domain.ErrInvalidOrderStatus)
	}

	return order, nil
}

func domainOrderToModel(order domain.Order) orderModel {
	model := orderModel{
		ID:         order.ID,
		UserID:     order.UserID,
		PartIDs:    order.PartIDs,
		TotalPrice: order.TotalPrice,
		Status:     string(order.Status),
	}

	if order.TransactionID != nil {
		model.TransactionID = order.TransactionID
	}

	if order.PaymentMethod != nil {
		model.PaymentMethod = sql.NullString{String: string(*order.PaymentMethod), Valid: true}
	}

	return model
}
