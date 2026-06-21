package controllers

import (
	"net/http"
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
