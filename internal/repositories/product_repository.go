package repositories

import (
	"context"
	"product-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductRepository struct {
	Collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) *ProductRepository {
	return &ProductRepository{
		Collection: db.Collection("products"),
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	product.ID = primitive.NewObjectID()
	_, err := r.Collection.InsertOne(ctx, product)
	return err
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*models.Product, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var product models.Product
	err = r.Collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *ProductRepository) FindAll(ctx context.Context) ([]*models.Product, error) {
	cursor, err := r.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*models.Product
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, id string, updateData *models.Product) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": updateData,
	}

	_, err = r.Collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.Collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}
