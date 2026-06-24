package main

import (
	"log"
	"os"
	"time"
	"product-service/internal/config"
	"product-service/internal/controllers"
	"product-service/internal/repositories"
	"product-service/internal/services"
	"product-service/internal/middlewares"

	"github.com/gin-gonic/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Khởi tạo kết nối DB và Cache
	config.ConnectDB()
	config.ConnectRedis()

	// 2. Khởi tạo các layer theo Dependency Injection
	repo := repositories.NewProductRepository(config.DB)
	service := services.NewProductService(repo, config.RedisClient)
	controller := controllers.NewProductController(service)

	// 3. Khởi tạo Gin Router
	router := gin.Default()

	// CẤU HÌNH CORS CHO FRONTEND
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://frontend-soa-gray.vercel.app"}, // Chỉ cho phép link này gọi API
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/uploads", "./uploads")

	// 4. Định nghĩa Routes
	// 4.1. PUBLIC API: Ai cũng xem được danh sách sản phẩm (Không cần middleware)
	publicAPI := router.Group("/api/v1")
	{
		publicAPI.GET("/categories", controller.GetAllCategories)

		publicAPI.GET("/products", controller.GetAllProducts)
		publicAPI.GET("/products/flash-sale", controller.GetFlashSaleProducts)
		publicAPI.GET("/products/:id", controller.GetProduct)
	}

	// 4.2. ADMIN API: Bắt buộc đăng nhập (RequireAuth) VÀ phải là Admin (RequireAdmin)
	adminAPI := router.Group("/api/v1")
	adminAPI.Use(middlewares.RequireAuth(), middlewares.RequireAdmin())
	{
		adminAPI.POST("/products", controller.CreateProduct)
		adminAPI.PUT("/products/:id", controller.UpdateProduct)
		adminAPI.DELETE("/products/:id", controller.DeleteProduct)
		adminAPI.POST("/products/:id/upload", controller.UploadProductImage)
	}

	// 4.3. INTERNAL / USER API: Các thao tác mua bán
	userAPI := router.Group("/api/v1/internal")
	userAPI.Use(middlewares.RequireAuth()) // Bắt buộc đăng nhập (User hay Admin đều được)
	{
		userAPI.PUT("/products/bulk-stock", controller.BulkUpdateStock) // Hàm nhận mảng ID
	}

	// 5. CẤU HÌNH CỔNG ĐỘNG CHO RENDER
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Nếu chạy ở máy tính bạn thì vẫn dùng 8082
	}

	log.Printf("Server đang khởi chạy tại port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Lỗi khởi chạy server:", err)
	}
}
