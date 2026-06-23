package main

import (
	"log"
	"product-service/internal/config"
	"product-service/internal/controllers"
	"product-service/internal/repositories"
	"product-service/internal/services"

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
	api := router.Group("/api/v1")
	{
		api.POST("/products", controller.CreateProduct)
		api.GET("/products", controller.GetAllProducts)

		api.GET("/products/flash-sale", controller.GetFlashSaleProducts)

		api.GET("/products/:id", controller.GetProduct)
		api.PUT("/products/:id", controller.UpdateProduct)
		api.DELETE("/products/:id", controller.DeleteProduct)

		api.POST("/products/:id/upload", controller.UploadProductImage)
	}

	// 5. Chạy server ở cổng 8082 như chốt ban đầu
	log.Println("Server đang chạy tại http://localhost:8082")
	if err := router.Run(":8082"); err != nil {
		log.Fatal("Lỗi khởi chạy server:", err)
	}
}
