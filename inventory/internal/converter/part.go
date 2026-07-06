package converter

import (
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/horizoonn/factory-platform/inventory/internal/domain"
	inventorypb "github.com/horizoonn/factory-platform/shared/pkg/proto/inventory/v1"
)

func PartsToProto(parts []domain.Part) []*inventorypb.Part {
	partsProto := make([]*inventorypb.Part, 0, len(parts))
	for _, part := range parts {
		partsProto = append(partsProto, PartToProto(&part))
	}

	return partsProto
}

func PartToProto(part *domain.Part) *inventorypb.Part {
	var meta map[string]*inventorypb.Value
	if part.Metadata != nil {
		meta = make(map[string]*inventorypb.Value, len(part.Metadata))
		for k, v := range part.Metadata {
			meta[k] = ValueToProto(v)
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

	var dimensionsProto *inventorypb.Dimensions
	if part.Dimensions != nil {
		dimensionsProto = &inventorypb.Dimensions{
			Length: part.Dimensions.Length,
			Width:  part.Dimensions.Width,
			Height: part.Dimensions.Height,
			Weight: part.Dimensions.Weight,
		}
	}

	var manufacturerProto *inventorypb.Manufacturer
	if part.Manufacturer != nil {
		manufacturerProto = &inventorypb.Manufacturer{
			Name:    part.Manufacturer.Name,
			Country: part.Manufacturer.Country,
			Website: part.Manufacturer.Website,
		}
	}

	return &inventorypb.Part{
		Uuid:          part.ID.String(),
		Name:          part.Name,
		Description:   part.Description,
		Price:         part.Price,
		StockQuantity: part.StockQuantity,
		Category:      inventorypb.Category(int32(part.Category)), //nolint:gosec
		Dimensions:    dimensionsProto,
		Manufacturer:  manufacturerProto,
		Tags:          part.Tags,
		Metadata:      meta,
		CreatedAt:     createdAtProto,
		UpdatedAt:     updatedAtProto,
	}
}

func ValueToProto(v *domain.Value) *inventorypb.Value {
	if v == nil {
		return &inventorypb.Value{}
	}
	switch {
	case v.String != nil:
		return &inventorypb.Value{
			Kind: &inventorypb.Value_StringValue{
				StringValue: *v.String,
			},
		}
	case v.Int64 != nil:
		return &inventorypb.Value{
			Kind: &inventorypb.Value_Int64Value{
				Int64Value: *v.Int64,
			},
		}
	case v.Float64 != nil:
		return &inventorypb.Value{
			Kind: &inventorypb.Value_DoubleValue{
				DoubleValue: *v.Float64,
			},
		}
	case v.Bool != nil:
		return &inventorypb.Value{
			Kind: &inventorypb.Value_BoolValue{
				BoolValue: *v.Bool,
			},
		}
	default:
		return &inventorypb.Value{}
	}
}

func FilterToDomain(filter *inventorypb.PartsFilter) (domain.PartsFilter, error) {
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
