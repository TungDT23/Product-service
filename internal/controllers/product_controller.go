package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"product-service/internal/models"
	"product-service/internal/services"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	Service *services.ProductService
}

func NewProductController(service *services.ProductService) *ProductController {
	return &ProductController{
		Service: service,
	}
}

func (c *ProductController) CreateProduct(ctx *gin.Context) {
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.Status = "active"
	
	err := c.Service.CreateProduct(ctx.Request.Context(), &product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể tạo sản phẩm"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Tạo sản phẩm thành công", "product_id": product.ID})
}

func (c *ProductController) GetProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	
	// Đo thời gian để Demo cho thầy giáo thấy sự khác biệt của Redis
	startTime := time.Now()

	product, err := c.Service.GetProductByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Sản phẩm không tồn tại"})
		return
	}

	duration := time.Since(startTime).Milliseconds()

	ctx.JSON(http.StatusOK, gin.H{
		"data": product,
		"meta": gin.H{
			"response_time_ms": duration, // Sẽ show ra ở Postman
		},
	})
}

func (c *ProductController) GetAllProducts(ctx *gin.Context) {
	products, err := c.Service.GetAllProducts(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách sản phẩm"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": products})
}

func (c *ProductController) UpdateProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.Service.UpdateProduct(ctx.Request.Context(), id, &product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật sản phẩm"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cập nhật sản phẩm thành công"})
}

func (c *ProductController) DeleteProduct(ctx *gin.Context) {
	id := ctx.Param("id")
	
	err := c.Service.DeleteProduct(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi xóa sản phẩm"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Xóa sản phẩm thành công"})
}

func (c *ProductController) UploadProductImage(ctx *gin.Context) {
	productID := ctx.Param("id")

	// 1. Nhận file từ form (key là "image")
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file ảnh trong request"})
		return
	}

	// 2. Tạo tên file duy nhất để tránh bị trùng đè (thêm timestamp)
	extension := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s_%d%s", productID, time.Now().Unix(), extension)
	
	// Đường dẫn lưu file vật lý trên server (Đảm bảo bạn đã tạo thư mục "uploads" ở thư mục gốc)
	savePath := filepath.Join("uploads", newFileName)

	// 3. Lưu file vào ổ cứng của Server
	if err := ctx.SaveUploadedFile(file, savePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lưu file"})
		return
	}

	// 4. Tạo URL để Frontend có thể truy cập được ảnh
	// Giả sử server Go chạy ở localhost:8082
	imageUrl := fmt.Sprintf("http://localhost:8082/uploads/%s", newFileName)

	// 5. Gọi Service để lưu URL này vào MongoDB
	err = c.ProductService.UploadImage(ctx.Request.Context(), productID, imageUrl)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Lưu URL vào database thất bại"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Upload ảnh thành công",
		"image_url": imageUrl,
	})
}