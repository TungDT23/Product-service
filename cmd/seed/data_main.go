package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"product-service/internal/config"
	"product-service/internal/models"

	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Hàm hỗ trợ lấy dữ liệu từ cột an toàn (tránh lỗi out of index nếu ô trống)
func getCell(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

func main() {
	// 1. Tải cấu hình và kết nối Database
	if err := godotenv.Load(); err != nil {
		log.Println("Cảnh báo: Không tìm thấy file .env, dùng biến môi trường OS")
	}
	config.ConnectDB()
	db := config.DB

	// THÊM 2 DÒNG NÀY ĐỂ DEBUG:
	fmt.Println("🔍 ĐƯỜNG LINK ĐANG DÙNG:", os.Getenv("MONGO_URI"))
	fmt.Println("🔍 TÊN DATABASE ĐANG DÙNG:", db.Name())

	// 2. Dọn rác Database cũ (Reset)
	log.Println("Đang xóa dữ liệu cũ...")
	db.Collection("categories").Drop(context.Background())
	db.Collection("products").Drop(context.Background())

	// 3. Mở file Excel
	f, err := excelize.OpenFile("Data.xlsx")
	if err != nil {
		log.Fatal("Lỗi mở file Excel (Kiểm tra lại tên file có đúng là Data.xlsx không): ", err)
	}
	defer f.Close()

	// -----------------------------------------
	// 4. XỬ LÝ SHEET CATEGORIES
	// -----------------------------------------
	catRows, err := f.GetRows("Categories")
	if err != nil {
		log.Fatal("Lỗi đọc sheet Categories: ", err)
	}

	var categories []interface{}
	for i, row := range catRows {
		if i == 0 || len(row) == 0 || getCell(row, 0) == "" {
			continue // Bỏ qua dòng tiêu đề hoặc dòng trống
		}

		cat := models.Category{
			ID:   primitive.NewObjectID(),
			Slug: getCell(row, 0),
			Name: getCell(row, 1), // Tên hiển thị (cột 1)
		}
		categories = append(categories, cat)
	}

	if len(categories) > 0 {
		_, err = db.Collection("categories").InsertMany(context.Background(), categories)
		if err != nil {
			log.Fatal("Lỗi thêm Categories: ", err)
		}
		fmt.Printf("✅ Đã thêm %d danh mục!\n", len(categories))
	}

	// -----------------------------------------
	// 5. XỬ LÝ SHEET PRODUCTS
	// -----------------------------------------
	prodRows, err := f.GetRows("Products")
	if err != nil {
		log.Fatal("Lỗi đọc sheet Products: ", err)
	}

	var products []interface{}
	for i, row := range prodRows {
		if i == 0 || len(row) == 0 || getCell(row, 0) == "" {
			continue // Bỏ qua dòng tiêu đề hoặc dòng không có tên
		}

		// Parse các trường dạng số
		price, _ := strconv.ParseFloat(getCell(row, 2), 64)
		stock, _ := strconv.Atoi(getCell(row, 3))
		sold, _ := strconv.Atoi(getCell(row, 4))
		discountPrice, _ := strconv.ParseFloat(getCell(row, 5), 64)
		discountPercent, _ := strconv.Atoi(getCell(row, 6))

		// Xử lý mảng ảnh (cắt theo dấu phẩy) - Cột 12
		rawImages := getCell(row, 12)
		var images []string
		if rawImages != "" {
			parts := strings.Split(rawImages, ",")
			for _, p := range parts {
				if img := strings.TrimSpace(p); img != "" {
					images = append(images, img)
				}
			}
		}

		// Xử lý JSON cho trường Attributes - Cột 13
		rawAttr := getCell(row, 13)
		var attributes map[string]interface{}
		if rawAttr != "" {
			if err := json.Unmarshal([]byte(rawAttr), &attributes); err != nil {
				log.Printf("⚠️ Cảnh báo: Dòng %d (Sản phẩm: %s) sai định dạng JSON ở cột Attributes_JSON. Sẽ bỏ qua thuộc tính này.\n", i+1, getCell(row, 0))
			}
		}

		// Xử lý Status mặc định - Cột 14
		status := getCell(row, 14)
		if status == "" {
			status = "active"
		}

		// Khởi tạo Product model
		prod := models.Product{
			ID:              primitive.NewObjectID(),
			Name:            getCell(row, 0),
			Description:     getCell(row, 1),
			Price:           price,
			Stock:           stock,
			Sold:            sold,
			DiscountPrice:   discountPrice,
			DiscountPercent: discountPercent,
			CategoryID:      getCell(row, 9),
			VendorID:        getCell(row, 10),
			Brand:           getCell(row, 11),
			Images:          images,
			Attributes:      attributes,
			Status:          status,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		// Xử lý Flash Sale Date nếu có điền trong Excel - Cột 7 và 8
		startDateStr := getCell(row, 7)
		endDateStr := getCell(row, 8)
		if startDateStr != "" && endDateStr != "" {
			if start, err := time.Parse("2006-01-02", startDateStr); err == nil {
				prod.SaleStartDate = &start
			}
			if end, err := time.Parse("2006-01-02", endDateStr); err == nil {
				prod.SaleEndDate = &end
			}
		}

		products = append(products, prod)
	}

	if len(products) > 0 {
		_, err = db.Collection("products").InsertMany(context.Background(), products)
		if err != nil {
			log.Fatal("Lỗi thêm Products: ", err)
		}
		fmt.Printf("✅ Đã thêm %d sản phẩm thành công!\n", len(products))
	}

	fmt.Println("🚀 HOÀN TẤT BƠM DỮ LIỆU!")
}