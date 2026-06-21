package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Description string                 `bson:"description" json:"description"`
	Price       float64                `bson:"price" json:"price"`
	Stock       int                    `bson:"stock" json:"stock"`
	CategoryID  string                 `bson:"category_id" json:"category_id"`
	VendorID    string                 `bson:"vendor_id" json:"vendor_id"`
	Brand       string                 `bson:"brand" json:"brand"`
	Images      []string               `bson:"images" json:"images"`
	Attributes  map[string]interface{} `bson:"attributes" json:"attributes"`
	Status      string                 `bson:"status" json:"status"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}
