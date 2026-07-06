package domain

import "errors"

var (
	ErrNotFound             = errors.New("not found")
	ErrNotImplemented       = errors.New("not implemented")
	ErrOrderIDRequired      = errors.New("order id is required")
	ErrUserIDRequired       = errors.New("user id is required")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrOrderAlreadyPaid     = errors.New("order already paid")
	ErrOrderCancelled       = errors.New("order is cancelled")
	ErrPartsNotFound        = errors.New("some parts not found")
	ErrEmptyParts           = errors.New("parts list is empty")
)
