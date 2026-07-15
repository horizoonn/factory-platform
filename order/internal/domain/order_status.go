package domain

type OrderStatus string

const (
	OrderStatusUnknown        OrderStatus = ""
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCompleted      OrderStatus = "COMPLETED"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
)

func (s OrderStatus) Valid() bool {
	switch s {
	case OrderStatusPendingPayment,
		OrderStatusPaid,
		OrderStatusCompleted,
		OrderStatusCancelled:
		return true
	default:
		return false
	}
}
