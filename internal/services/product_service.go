package services

import (
	"context"
	"encoding/json"
	"fmt"
	"errors"
	"product-service/internal/models"
	"product-service/internal/repositories"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductService struct {
	Repo        *repositories.ProductRepository
	RedisClient *redis.Client
}

func NewProductService(repo *repositories.ProductRepository, redisClient *redis.Client) *ProductService {
	return &ProductService{
		Repo:        repo,
		RedisClient: redisClient,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	// BẮT BUỘC Category phải tồn tại từ trước
	if product.CategoryID != "" {
		exists, err := s.Repo.CheckCategoryExists(ctx, product.CategoryID)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("danh mục không tồn tại, vui lòng chọn danh mục hợp lệ")
		}
	} else {
		return errors.New("category_id không được để trống")
	}

	// Nếu mọi thứ hợp lệ, mới lưu sản phẩm vào DB
	return s.Repo.Create(ctx, product)
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	// 1. Kiểm tra cache Redis trước (Cache-Aside Pattern)
	cacheKey := fmt.Sprintf("product:%s", id)
	cachedProduct, err := s.RedisClient.Get(ctx, cacheKey).Result()
	
	if err == nil {
		// Cache Hit
		var product models.Product
		if err := json.Unmarshal([]byte(cachedProduct), &product); err == nil {
			return &product, nil
		}
	}

	// 2. Cache Miss - Lấy từ MongoDB
	product, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Lưu vào Redis cho lần sau (TTL: 10 phút)
	productJSON, _ := json.Marshal(product)
	s.RedisClient.Set(ctx, cacheKey, productJSON, 10*time.Minute)

	return product, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context, limit int64, skip int64, search string, category string, minPrice float64, maxPrice float64) ([]*models.Product, error) {
	// Truyền toàn bộ tham số lọc xuống Repository
	return s.Repo.FindAll(ctx, limit, skip, search, category, minPrice, maxPrice)
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, product *models.Product) error {
	err := s.Repo.Update(ctx, id, product)
	if err != nil {
		return err
	}

	// Cache Invalidation: Xóa cache cũ đi để lần tới user xem sẽ lấy data mới nhất từ MongoDB
	cacheKey := fmt.Sprintf("product:%s", id)
	s.RedisClient.Del(ctx, cacheKey)

	return nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	// 1. Lấy thông tin sản phẩm ra trước (Giả sử bạn đã có hàm GetProductByID trong Repo)
	product, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Thực hiện xóa sản phẩm (Code cũ của bạn)
	err = s.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// 💡 3. Kiểm tra dọn dẹp Category
	if product.CategoryID != "" {
		count, _ := s.Repo.CountProductsByCategory(ctx, product.CategoryID)
		if count == 0 { // Hết sạch sản phẩm -> Tiêu diệt danh mục
			_ = s.Repo.DeleteCategoryByName(ctx, product.CategoryID)
		}
	}

	return nil
}

func (s *ProductService) GetFlashSaleProducts(ctx context.Context, limit int64) ([]*models.Product, error) {
	return s.Repo.FindFlashSales(ctx, limit)
}

func (s *ProductService) UpdateStockAndSold(ctx context.Context, id string, quantity int) error {
	return s.Repo.UpdateStockAndSold(ctx, id, quantity)
}

func (s *ProductService) BulkUpdateStock(ctx context.Context, items map[string]int) error {
	return s.Repo.BulkUpdateStock(ctx, items)
}

func (s *ProductService) GetAllWithoutPagination(ctx context.Context) ([]*models.Product, error) {
    // Tên biến Repo ở đây tùy thuộc vào struct Service của bạn (ví dụ s.repo hoặc s.productRepo)
    return s.Repo.GetAllWithoutPagination(ctx) 
}