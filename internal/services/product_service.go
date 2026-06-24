package services

import (
	"context"
	"encoding/json"
	"fmt"
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
	err := s.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	// Cache Invalidation
	cacheKey := fmt.Sprintf("product:%s", id)
	s.RedisClient.Del(ctx, cacheKey)

	return nil
}

func (s *ProductService) UploadImage(ctx context.Context, id string, imageUrl string) error {
	// 1. Lấy sản phẩm hiện tại ra
	product, err := s.Repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Thêm URL mới vào mảng Images
	product.Images = append(product.Images, imageUrl)
	product.UpdatedAt = time.Now()

	// 3. Cập nhật lại vào MongoDB
	err = s.Repo.Update(ctx, id, product)
	if err != nil {
		return err
	}

	// 4. Xóa Cache Redis để dữ liệu đồng bộ
	cacheKey := fmt.Sprintf("product:%s", id)
	s.RedisClient.Del(ctx, cacheKey)

	return nil
}

func (s *ProductService) GetFlashSaleProducts(ctx context.Context, limit int64) ([]*models.Product, error) {
	return s.Repo.FindFlashSales(ctx, limit)
}

func (s *ProductService) UpdateStockAndSold(ctx context.Context, id string, quantity int) error {
	return s.Repo.UpdateStockAndSold(ctx, id, quantity)
}