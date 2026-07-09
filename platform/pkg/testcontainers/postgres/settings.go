package postgres

import "time"

const (
	DefaultOperationTimeout = 5 * time.Second
	DefaultMaxConnIdleTime  = time.Minute
	DefaultMaxConns         = int32(4)
	DefaultMinConns         = int32(1)
)
