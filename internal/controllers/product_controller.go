package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"product-service/internal/models"
	"product-service/internal/services"
	"product-service/internal/repositories"
	"time"
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	Service *services.ProductService
	productRepo *repositories.ProductRepository
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
	// Lấy tham số phân trang
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	page, _ := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
	skip := (page - 1) * limit

	// HỨNG CÁC THAM SỐ LỌC TỪ URL
	search := ctx.Query("search")
	category := ctx.Query("category_id")
	minPrice, _ := strconv.ParseFloat(ctx.DefaultQuery("min_price", "0"), 64)
	maxPrice, _ := strconv.ParseFloat(ctx.DefaultQuery("max_price", "0"), 64)

	// Truyền dữ liệu xuống Service
	products, err := c.Service.GetAllProducts(ctx.Request.Context(), limit, skip, search, category, minPrice, maxPrice)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"page":  page,
		"limit": limit,
		"data":  products,
	})
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

func (c *ProductController) UploadImage(ctx *gin.Context) {
	// 1. Phân tích toàn bộ form dữ liệu gửi lên
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Lỗi định dạng form dữ liệu"})
		return
	}

	// 2. Lấy danh sách các file từ key "images" (Lưu ý: đã đổi thành số nhiều)
	files := form.File["images"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Không tìm thấy file ảnh nào"})
		return
	}

	// Chuẩn bị các biến để tạo đường dẫn
	scheme := "http"
	if ctx.Request.TLS != nil || ctx.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := ctx.Request.Host

	var imageUrls []string // Mảng chứa các link ảnh sau khi lưu thành công

	// 3. Vòng lặp xử lý từng file
	for _, file := range files {
		// Tạo tên file độc nhất, có thể cộng thêm chuỗi ngẫu nhiên để không bị trùng nếu up file quá nhanh
		newFileName := fmt.Sprintf("img_%d_%s", time.Now().UnixNano(), file.Filename)
		savePath := filepath.Join("uploads", newFileName)

		// Lưu file vào ổ cứng
		if err := ctx.SaveUploadedFile(file, savePath); err != nil {
			// Nếu 1 file lỗi, báo log và bỏ qua file đó, chạy tiếp file sau
			fmt.Println("Lỗi lưu file:", err)
			continue
		}

		// Ghép thành link hoàn chỉnh và nhét vào mảng
		url := fmt.Sprintf("%s://%s/uploads/%s", scheme, host, newFileName)
		imageUrls = append(imageUrls, url)
	}

	// 4. Trả về mảng các đường link
	ctx.JSON(http.StatusOK, gin.H{
		"message":    "Upload ảnh thành công",
		"image_urls": imageUrls, // Trả về dạng mảng []string
	})
}

func (c *ProductController) GetFlashSaleProducts(ctx *gin.Context) {
	// Trang chủ thường chỉ lấy giới hạn số lượng (vd: top 10 sản phẩm)
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)

	products, err := c.Service.GetFlashSaleProducts(ctx.Request.Context(), limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy dữ liệu Flash Sale"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"limit": limit,
		"data":  products,
	})
}

func (c *ProductController) UpdateStock(ctx *gin.Context) {
	id := ctx.Param("id")

	// Tạo một struct ẩn (anonymous struct) để hứng dữ liệu quantity từ JSON
	var requestBody struct {
		Quantity int `json:"quantity" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ hoặc thiếu số lượng (quantity)"})
		return
	}

	err := c.Service.UpdateStockAndSold(ctx.Request.Context(), id, requestBody.Quantity)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cập nhật kho và doanh số thành công"})
}

type OrderItem struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

func (c *ProductController) BulkUpdateStock(ctx *gin.Context) {
	var requestBody struct {
		Items []OrderItem `json:"items" binding:"required,dive"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu mảng sản phẩm không hợp lệ"})
		return
	}

	// Chuyển đổi mảng thành map map[id]quantity để ném xuống Service
	updateMap := make(map[string]int)
	for _, item := range requestBody.Items {
		updateMap[item.ProductID] = item.Quantity
	}

	err := c.Service.BulkUpdateStock(ctx.Request.Context(), updateMap)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi cập nhật kho hàng loạt"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cập nhật giỏ hàng thành công"})
}

func (ctrl *ProductController) GetAllProductsNoPagination(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. CHÚ Ý TẠI ĐÂY: Thay chữ "Repo" thành tên chuẩn của bạn.
	// Ví dụ: ctrl.productRepo.GetAllWithoutPagination(ctx)
	// Hoặc: ctrl.ProductService.GetAllWithoutPagination(ctx)
	products, err := ctrl.productRepo.GetAllWithoutPagination(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi lấy toàn bộ dữ liệu sản phẩm"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  products,
		"total": len(products),
	})
}