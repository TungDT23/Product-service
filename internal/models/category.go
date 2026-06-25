package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Category struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Slug string				`bson:"slug" json:"slug"`
	Name string             `bson:"name" json:"name"`
}