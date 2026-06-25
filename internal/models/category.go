package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Category struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CategoryID   string             `bson:"category_id" json:"category_id"` // Trả về "C001"
	CategorySlug string             `bson:"category_slug" json:"slug"`      // Trả về "smartphones"
	Name         string             `bson:"name" json:"name"`               // Trả về "Điện thoại di động"
}