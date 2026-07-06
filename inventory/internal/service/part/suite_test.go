package part

import (
	"errors"

	"github.com/google/uuid"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
)

var (
	partID        = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	secondPartID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	errRepository = errors.New("repository error")
)

func validPart() domain.Part {
	return domain.Part{
		ID:            partID,
		Name:          "engine",
		Description:   "main engine",
		Price:         100,
		StockQuantity: 10,
		Category:      domain.CATEGORY_ENGINE,
	}
}

func secondPart() domain.Part {
	part := validPart()
	part.ID = secondPartID
	part.Name = "wing"
	part.Description = "left wing"
	part.Price = 200
	part.StockQuantity = 5
	part.Category = domain.CATEGORY_WING
	return part
}

func validPartsFilter() domain.PartsFilter {
	return domain.PartsFilter{
		UUIDs: []uuid.UUID{partID, secondPartID},
	}
}
