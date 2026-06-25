package main

import (
	"log"
	"os"
	"time"

	"product-service/internal/config"
	"product-service/internal/controllers"
	"product-service/internal/middlewares"
	"product-service/internal/repositories"
	"product-service/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 0. Đọc file .env khi code ở máy tính (Lên Render không có file này nó sẽ tự báo cảnh báo nhưng vẫn chạy tiếp)
	if err := godotenv.Load(); err != nil {
		log.Println("Cảnh báo: Không tìm thấy file .env, hệ thống sẽ sử dụng biến môi trường của OS")
	}

	// 1. Khởi tạo kết nối DB và Cache
	config.ConnectDB()
	config.ConnectRedis()

	// 2. Khởi tạo các layer theo Dependency Injection
	repo := repositories.NewProductRepository(config.DB)
	service := services.NewProductService(repo, config.RedisClient)
	productController := controllers.NewProductController(service)
	categoryController := controllers.NewCategoryController(service)

	// 3. Khởi tạo Gin Router
	router := gin.Default()

	// CẤU HÌNH CORS CHO FRONTEND
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://frontend-soa-gray.vercel.app", "http://localhost:5173"}, // Chỉ cho phép link này gọi API
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/uploads", "./uploads")

	// 4. Định nghĩa Routes
	// 4.1. PUBLIC API: Ai cũng xem được danh sách sản phẩm
	publicAPI := router.Group("/api/v1")
	{
		publicAPI.GET("/categories", categoryController.GetAllCategories)

		publiCAPI.GET("/products/all", productController.GetAllProductsNoPagination)

		publicAPI.GET("/products", productController.GetAllProducts)
		publicAPI.GET("/products/flash-sale", productController.GetFlashSaleProducts)
		publicAPI.GET("/products/:id", productController.GetProduct)
	}

	// 4.2. ADMIN API: Bắt buộc đăng nhập VÀ phải là Admin
	adminAPI := router.Group("/api/v1")
	adminAPI.Use(middlewares.RequireAuth(), middlewares.RequireAdmin())
	{
		adminAPI.POST("/products", productController.CreateProduct)
		adminAPI.PUT("/products/:id", productController.UpdateProduct)
		adminAPI.DELETE("/products/:id", productController.DeleteProduct)

		adminAPI.POST("/upload", productController.UploadImage)
	}

	// 4.3. INTERNAL / USER API: Các thao tác mua bán
	userAPI := router.Group("/api/v1/internal")
	userAPI.Use(middlewares.RequireAuth())
	{
		userAPI.PUT("/products/bulk-stock", productController.BulkUpdateStock)
	}

	// 5. CẤU HÌNH CỔNG ĐỘNG CHO RENDER
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Dành cho máy local
	}

	log.Printf("Server đang khởi chạy tại port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Lỗi khởi chạy server:", err)
	}
}