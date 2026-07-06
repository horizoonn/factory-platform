package inventoryv1

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/client/converter"
	"github.com/horizoonn/factory-platform/order/internal/domain"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func (c *Client) listParts(ctx context.Context, partIDs []uuid.UUID) ([]domain.Part, error) {
	ids := make([]string, 0, len(partIDs))
	for _, partID := range partIDs {
		ids = append(ids, partID.String())
	}

	resp, err := c.client.ListParts(ctx, &inventorypb.ListPartsRequest{
		Filter: &inventorypb.PartsFilter{
			Uuids: ids,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list parts grpc call: %w", err)
	}

	parts := make([]domain.Part, 0, len(resp.GetParts()))
	for _, part := range resp.GetParts() {
		convertedPart, err := converter.PartFromProto(part)
		if err != nil {
			return nil, fmt.Errorf("convert part from proto: %w", err)
		}

		parts = append(parts, convertedPart)
	}

	return parts, nil
}
