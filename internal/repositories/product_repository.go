package repositories

import (
	"errors"
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
	CategoryCollection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) *ProductRepository {
	return &ProductRepository{
		Collection: db.Collection("products"),
		CategoryCollection: db.Collection("categories"),
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

func (r *ProductRepository) FindAll(ctx context.Context, limit int64, skip int64, search string, category string, minPrice float64, maxPrice float64) ([]*models.Product, error) {
    // Sửa 1: Khởi tạo mảng sẵn để trả về "[]" thay vì "null" nếu không có sản phẩm
    products := make([]*models.Product, 0)

    filter := bson.M{"status": "active"}

    // 2. Lọc theo tên (Tìm kiếm gần đúng Regex, chữ "i" là không phân biệt hoa thường)
    if search != "" {
        filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}
    }

    // 3. Lọc theo danh mục (Hỗ trợ cả ID loằng ngoằng và chữ "smartphones")
	if category != "" {
		if primitive.IsValidObjectID(category) {
			// Nếu Frontend truyền ID -> Lọc theo cột category_id
			filter["category_id"] = category
		} else {
			// Nếu Frontend truyền chữ -> Lọc theo cột category_slug
			filter["category_slug"] = category
		}
	}

    // 4. Lọc theo khoảng giá
    if minPrice > 0 || maxPrice > 0 {
        priceFilter := bson.M{}
        if minPrice > 0 {
            priceFilter["$gte"] = minPrice // $gte: Lớn hơn hoặc bằng
        }
        if maxPrice > 0 {
            priceFilter["$lte"] = maxPrice // $lte: Nhỏ hơn hoặc bằng
        }
        filter["price"] = priceFilter
    }

    findOptions := options.Find()
    findOptions.SetLimit(limit)
    findOptions.SetSkip(skip)

    cursor, err := r.Collection.Find(ctx, filter, findOptions)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

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

func (r *ProductRepository) FindFlashSales(ctx context.Context, limit int64) ([]*models.Product, error) {
	var products []*models.Product
	now := time.Now()

	// 1. Điều kiện lọc: Đang sale (discount > 0) VÀ thời gian kết thúc > hiện tại
	filter := bson.M{
		"discount_percent": bson.M{"$gt": 0},
		"sale_end_date":    bson.M{"$gt": now},
		"status":           "active",
	}

	// 2. Tùy chọn truy vấn: Ưu tiên xếp sản phẩm giảm giá sâu lên đầu
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"discount_percent": -1}) // -1 là giảm dần (DESC)
	findOptions.SetLimit(limit)

	// 3. Thực thi truy vấn
	cursor, err := r.Collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 4. Đọc dữ liệu
	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}

func (r *ProductRepository) UpdateStockAndSold(ctx context.Context, id string, quantity int) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	// Dùng $inc (increment) để cộng/trừ thẳng trên DB, chống lỗi nhiều người mua cùng lúc
	update := bson.M{
		"$inc": bson.M{
			"stock": -quantity, // Trừ đi số lượng khách mua
			"sold":  quantity,  // Cộng vào số lượng đã bán
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	result, err := r.Collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return errors.New("không tìm thấy sản phẩm")
	}

	return nil
}

func (r *ProductRepository) BulkUpdateStock(ctx context.Context, items map[string]int) error {
	var models []mongo.WriteModel

	for id, quantity := range items {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err // Báo lỗi nếu có ID nào không hợp lệ
		}

		// Tạo lệnh cập nhật cho từng sản phẩm
		update := bson.M{
			"$inc": bson.M{
				"stock": -quantity,
				"sold":  quantity,
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": objID}).
			SetUpdate(update)

		models = append(models, model)
	}

	// Chạy toàn bộ mảng lệnh trong 1 lần
	if len(models) > 0 {
		_, err := r.Collection.BulkWrite(ctx, models)
		return err
	}
	return nil
}

// 1. Lấy toàn bộ danh sách Category (Đã chuẩn hóa dùng Model)
func (r *ProductRepository) GetAllCategories(ctx context.Context) ([]*models.Category, error) {
	cursor, err := r.CategoryCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	
	var categories []*models.Category
	if err = cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	
	return categories, nil
}

// 2. Kiểm tra Category có tồn tại không (Thay cho hàm EnsureCategoryExists cũ)
func (r *ProductRepository) CheckCategoryExists(ctx context.Context, categoryID string) (bool, error) {
	// Giả sử bảng categories lưu ID dưới dạng text (name)
	count, err := r.CategoryCollection.CountDocuments(ctx, bson.M{"name": categoryID})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// 3. Đếm số sản phẩm trong 1 Category (Dùng lúc Xóa SP)
func (r *ProductRepository) CountProductsByCategory(ctx context.Context, categoryName string) (int64, error) {
	return r.Collection.CountDocuments(ctx, bson.M{"category_id": categoryName})
}

// 4. Xóa Category (Dùng lúc Xóa SP)
func (r *ProductRepository) DeleteCategoryByName(ctx context.Context, categoryName string) error {
	_, err := r.CategoryCollection.DeleteOne(ctx, bson.M{"name": categoryName})
	return err
}

// Lấy toàn bộ sản phẩm không phân trang
func (r *ProductRepository) GetAllWithoutPagination(ctx context.Context) ([]*models.Product, error) {
	// Khởi tạo mảng rỗng để tránh trả về null
	products := make([]*models.Product, 0)

	// Có thể chỉ lấy sản phẩm active, hoặc bỏ qua filter để lấy hết
	filter := bson.M{"status": "active"}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product models.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, nil
}