package inventoryv1

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

type Client struct {
	client inventorypb.InventoryServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{
		client: inventorypb.NewInventoryServiceClient(conn),
	}
}

func (c *Client) ListParts(ctx context.Context, partIDs []uuid.UUID) ([]domain.Part, error) {
	return c.listParts(ctx, partIDs)
}
