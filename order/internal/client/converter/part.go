package converter

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/order/internal/domain"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func PartFromProto(part *inventorypb.Part) (domain.Part, error) {
	if part == nil {
		return domain.Part{}, fmt.Errorf("part is nil")
	}

	id, err := uuid.Parse(part.GetUuid())
	if err != nil {
		return domain.Part{}, fmt.Errorf("parse part uuid: %w", err)
	}

	return domain.Part{
		ID:    id,
		Price: part.GetPrice(),
	}, nil
}
