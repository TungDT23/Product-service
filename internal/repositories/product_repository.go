package repositories

import (
	"context"
	"time"
	"product-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
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

func (r *ProductRepository) FindAll(ctx context.Context, limit int64, skip int64) ([]*models.Product, error) {
    findOptions := options.Find()
    findOptions.SetLimit(limit)
    findOptions.SetSkip(skip)

    cursor, err := r.Collection.Find(ctx, bson.M{}, findOptions)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var products []*models.Product
    if err = cursor.All(ctx, &products); err != nil {
        return nil, err
    }
    return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, id string, updateData *models.Product) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updateData.UpdatedAt = time.Now()

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
