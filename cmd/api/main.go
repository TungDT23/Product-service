package main

import (
	"log"
	"product-service/internal/config"
	"product-service/internal/controllers"
	"product-service/internal/repositories"
	"product-service/internal/services"
	"product-service/internal/middlewares"

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

	// 5. Chạy server ở cổng 8082 như chốt ban đầu
	log.Println("Server đang chạy tại http://localhost:8082")
	if err := router.Run(":8082"); err != nil {
		log.Fatal("Lỗi khởi chạy server:", err)
	}
}
