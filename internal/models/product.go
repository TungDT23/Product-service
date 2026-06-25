package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name" binding:"required"`
	Description string                 `bson:"description" json:"description"`
	Price       float64                `bson:"price" json:"price" binding:"required, min=0"`
	Stock       int                    `bson:"stock" json:"stock" binding:"min=0"`
	Sold		int					   `bson:"sold" json:"sold"`
	
	// --- Cho Flash Sale ---
	DiscountPrice   float64                `bson:"discount_price" json:"discount_price"`
	DiscountPercent int                    `bson:"discount_percent" json:"discount_percent"`
	SaleStartDate   *time.Time             `bson:"sale_start_date,omitempty" json:"sale_start_date"` // Dùng con trỏ để cho phép null
	SaleEndDate     *time.Time             `bson:"sale_end_date,omitempty" json:"sale_end_date"`     // Dùng con trỏ để cho phép null

	CategoryID  string                 `bson:"category_id" json:"category_id" binding:"required"`
	CategorySlug string				   `bson:"category_slug" json:"category_slug"`
	VendorID    string                 `bson:"vendor_id" json:"vendor_id"`
	Brand       string                 `bson:"brand" json:"brand"`
	Images      []string               `bson:"images" json:"images"`
	Attributes  map[string]interface{} `bson:"attributes" json:"attributes"`
	Status      string                 `bson:"status" json:"status"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}
