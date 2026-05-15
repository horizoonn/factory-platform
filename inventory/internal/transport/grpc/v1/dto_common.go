package grpcv1

import (
	"github.com/google/uuid"
	"github.com/horizoonn/factory-platform.git/inventory/internal/domain"
	inventoryv1 "github.com/horizoonn/factory-platform.git/shared/pkg/proto/inventory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func partsToProto(parts []domain.Part) []*inventoryv1.Part {
	partsProto := make([]*inventoryv1.Part, 0, len(parts))
	for _, part := range parts {
		partsProto = append(partsProto, partToProto(&part))
	}

	return partsProto
}

func partToProto(part *domain.Part) *inventoryv1.Part {
	var tags []string
	if part.Tags != nil {
		tags = part.Tags
	}

	var meta map[string]*inventoryv1.Value
	if part.Metadata != nil {
		meta = make(map[string]*inventoryv1.Value, len(part.Metadata))
		for k, v := range part.Metadata {
			meta[k] = valueToProto(v)
		}
	}

	var createdAtProto *timestamppb.Timestamp
	if part.CreatedAt != nil {
		createdAtProto = timestamppb.New(*part.CreatedAt)
	}

	var updatedAtProto *timestamppb.Timestamp
	if part.UpdatedAt != nil {
		updatedAtProto = timestamppb.New(*part.UpdatedAt)
	}

	var dimensionsProto *inventoryv1.Dimensions
	if part.Dimensions != nil {
		dimensionsProto = &inventoryv1.Dimensions{
			Length: part.Dimensions.Length,
			Width:  part.Dimensions.Width,
			Height: part.Dimensions.Height,
			Weight: part.Dimensions.Weight,
		}
	}

	var manufacturerProto *inventoryv1.Manufacturer
	if part.Manufacturer != nil {
		manufacturerProto = &inventoryv1.Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}

	return &inventoryv1.Part{
		Uuid:          part.ID.String(),
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      inventoryv1.Category(int32(part.Category)), //nolint:gosec
		Dimensions:    dimensionsProto,
		Manufacturer:  manufacturerProto,
		Tags:          tags,
		Metadata:      meta,
		CreatedAt:     createdAtProto,
		UpdatedAt:     updatedAtProto,
	}
}

func valueToProto(v *domain.Value) *inventoryv1.Value {
	if v == nil {
		return &inventoryv1.Value{}
	}
	switch {
	case v.String != nil:
		return &inventoryv1.Value{
			Kind: &inventoryv1.Value_StringValue{
				StringValue: *v.String,
			},
		}
	case v.Int64 != nil:
		return &inventoryv1.Value{
			Kind: &inventoryv1.Value_Int64Value{
				Int64Value: *v.Int64,
			},
		}
	case v.Float64 != nil:
		return &inventoryv1.Value{
			Kind: &inventoryv1.Value_DoubleValue{
				DoubleValue: *v.Float64,
			},
		}
	case v.Bool != nil:
		return &inventoryv1.Value{
			Kind: &inventoryv1.Value_BoolValue{
				BoolValue: *v.Bool,
			},
		}
	default:
		return &inventoryv1.Value{}
	}
}

func filterToDomain(filter *inventoryv1.PartsFilter) (domain.PartsFilter, error) {
	if filter == nil {
		return domain.PartsFilter{}, nil
	}

	ids := make([]uuid.UUID, 0, len(filter.GetUuids()))
	for _, rawID := range filter.GetUuids() {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return domain.PartsFilter{}, status.Errorf(codes.InvalidArgument, "invalid uuid: %s", rawID)
		}

		ids = append(ids, id)
	}

	categories := make([]domain.Category, 0, len(filter.GetCategories()))
	for _, category := range filter.GetCategories() {
		categories = append(categories, domain.Category(category))
	}

	return domain.PartsFilter{
		UUIDs:                 ids,
		Names:                 filter.GetNames(),
		Categories:            categories,
		ManufacturerCountries: filter.GetManufacturerCountries(),
		Tags:                  filter.GetTags(),
	}, nil
}
