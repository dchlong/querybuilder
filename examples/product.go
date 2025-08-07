package examples

import (
	"time"

	"gorm.io/datatypes"
)

// Product represents an e-commerce product with querybuilder annotation
//
//gen:querybuilder
type Product struct {
	ID          int64                           `json:"id"`
	Name        string                          `json:"name"`
	SKU         string                          `json:"sku"`
	Description *string                         `json:"description"`
	Price       float64                         `json:"price"`
	Stock       int                             `json:"stock"`
	CategoryID  int64                           `json:"category_id"`
	IsActive    bool                            `json:"is_active"`
	Tags        datatypes.JSONSlice[string]     `json:"tags"`       // JSON array
	Attributes  datatypes.JSONType[*Attributes] `json:"attributes"` // JSON object
	CreatedAt   time.Time                       `json:"created_at"`
	UpdatedAt   *time.Time                      `json:"updated_at"`
}

// Attributes represents product attributes stored in JSON
type Attributes struct {
	Color      string  `json:"color"`
	Size       string  `json:"size"`
	Weight     float64 `json:"weight"`
	Dimensions string  `json:"dimensions"`
}
