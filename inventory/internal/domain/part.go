package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category int32

const (
	CATEGORY_UNKNOWN  Category = 0
	CATEGORY_ENGINE   Category = 1
	CATEGORY_FUEL     Category = 2
	CATEGORY_PORTHOLE Category = 3
	CATEGORY_WING     Category = 4
)

type Part struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Price         float64
	StockQuantity int64
	Category      Category
	Dimensions    *Dimensions
	Manufacturer  *Manufacturer
	Tags          []string
	Metadata      map[string]*Value
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

func NewPart(
	id uuid.UUID,
	name string,
	description string,
	price float64,
	stockQuantity int64,
	category int32,
	dimensions *Dimensions,
	manufacturer *Manufacturer,
	tags []string,
	metadata map[string]*Value,
	createdAt *time.Time,
	updatedAt *time.Time,
) *Part {
	return &Part{
		ID:            id,
		Name:          name,
		Description:   description,
		Price:         price,
		StockQuantity: stockQuantity,
		Category:      Category(category),
		Dimensions:    dimensions,
		Manufacturer:  manufacturer,
		Tags:          tags,
		Metadata:      metadata,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

type Dimensions struct {
	Length float64
	Width  float64
	Height float64
	Weight float64
}

type Manufacturer struct {
	Name    string
	Country string
	Website string
}

type Value struct {
	String  *string
	Int64   *int64
	Float64 *float64
	Bool    *bool
}

type PartsFilter struct {
	UUIDs                 []uuid.UUID
	Names                 []string
	Categories            []Category
	ManufacturerCountries []string
	Tags                  []string
}
