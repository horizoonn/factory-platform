package domain

import "errors"

var (
	ErrNotFound       = errors.New("not found")
	ErrPartIDRequired = errors.New("part id is required")
)
