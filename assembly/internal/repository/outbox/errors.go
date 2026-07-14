package repository

import "errors"

var ErrLeaseLost = errors.New("outbox event lease lost")
