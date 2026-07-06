package converter

import (
	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	"github.com/horizoonn/factory-platform/inventory/internal/repository/model"
)

func PartModelToDomain(part model.Part) domain.Part {
	var dimensions *domain.Dimensions
	if part.Dimensions != nil {
		dimensions = &domain.Dimensions{
			Length: part.Dimensions.Length,
			Width:  part.Dimensions.Width,
			Height: part.Dimensions.Height,
			Weight: part.Dimensions.Weight,
		}
	}

	var manufacturer *domain.Manufacturer
	if part.Manufacturer != nil {
		manufacturer = &domain.Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}

	var metadata map[string]*domain.Value
	if part.Metadata != nil {
		metadata = make(map[string]*domain.Value, len(part.Metadata))
		for k, v := range part.Metadata {
			metadata[k] = valueToDomain(v)
		}
	}

	domainPart := domain.NewPart(
		part.ID,
		part.Name,
		part.Description,
		part.Price,
		part.StockQuantity,
		part.Category,
		dimensions,
		manufacturer,
		part.Tags,
		metadata,
		part.CreatedAt,
		part.UpdatedAt,
	)

	return *domainPart
}

func valueToDomain(v *model.Value) *domain.Value {
	if v == nil {
		return &domain.Value{}
	}
	return &domain.Value{
		String:  v.String,
		Int64:   v.Int64,
		Float64: v.Float64,
		Bool:    v.Bool,
	}
}
