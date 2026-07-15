package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	inventoryv1 "github.com/horizoonn/factory-platform/order/internal/client/grpc/inventory/v1"
	paymentv1 "github.com/horizoonn/factory-platform/order/internal/client/grpc/payment/v1"
)

type Clients struct {
	Inventory *inventoryv1.Client
	Payment   *paymentv1.Client
	conns     []*grpc.ClientConn
}

func NewClients(inventoryAddr, paymentAddr string) (*Clients, error) {
	inventoryConn, err := grpc.NewClient(
		inventoryAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create inventory grpc client: %w", err)
	}

	paymentConn, err := grpc.NewClient(
		paymentAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		if closeErr := inventoryConn.Close(); closeErr != nil {
			err = fmt.Errorf("close inventory connection: %w; original error: %w", closeErr, err)
		}
		return nil, fmt.Errorf("create payment grpc client: %w", err)
	}

	return &Clients{
		Inventory: inventoryv1.NewClient(inventoryConn),
		Payment:   paymentv1.NewClient(paymentConn),
		conns:     []*grpc.ClientConn{inventoryConn, paymentConn},
	}, nil
}

func (c *Clients) Close() error {
	var lastErr error
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (c *Clients) Connections() map[string]*grpc.ClientConn {
	return map[string]*grpc.ClientConn{
		"inventory": c.conns[0],
		"payment":   c.conns[1],
	}
}
