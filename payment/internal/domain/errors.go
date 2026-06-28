package domain

import "errors"

var (
	ErrInvalidPaymentMethod = errors.New("invalid payment method")
	ErrOrderIDRequired      = errors.New("order id is required")
	ErrUserIDRequired       = errors.New("user id is required")
)
