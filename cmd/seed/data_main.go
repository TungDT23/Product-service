package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"product-service/internal/config"
	"product-service/internal/models"

	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getCell(row []string, index int) string {
	if index < len(row) {
		return strings.TrimSpace(row[index])
	}
	return ""
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Cảnh báo: Không tìm thấy file .env")
	}
	config.ConnectDB()
	db := config.DB

	log.Println("Đang xóa dữ liệu cũ...")
	db.Collection("categories").Drop(context.Background())
	db.Collection("products").Drop(context.Background())

	f, err := excelize.OpenFile("Data.xlsx")
	if err != nil {
		log.Fatal("Lỗi mở file Excel: ", err)
	}
	defer f.Close()

	// ==========================================
	// 1. XỬ LÝ CATEGORIES & TẠO 2 TỪ ĐIỂN
	// ==========================================
	catRows, err := f.GetRows("Categories")
	if err != nil {
		log.Fatal("Lỗi đọc sheet Categories: ", err)
	}

	var categories []interface{}
	// Từ điển 1: Dịch "C001" -> "6a3cb5ad5ba..." (ID Mongo)
	categoryIDMap := make(map[string]string)
	// Từ điển 2: Dịch "C001" -> "smartphones" (Slug thân thiện)
	categorySlugMap := make(map[string]string)

	for i, row := range catRows {
		if i == 0 || len(row) == 0 || getCell(row, 0) == "" {
			continue // Bỏ qua dòng tiêu đề
		}

		catID := primitive.NewObjectID()
		excelCode := getCell(row, 0) // Cột A: Lấy mã "C001"
		slug := getCell(row, 1)      // Cột B: Lấy chữ "smartphones"

		// Ghi nhớ vào cả 2 từ điển
		categoryIDMap[excelCode] = catID.Hex()
		categorySlugMap[excelCode] = slug

		cat := models.Category{
			ID:           catID,
			CategoryID:   excelCode,
			CategorySlug: slug,
			Name:         getCell(row, 2), // Cột C: Tên danh mục
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

	// ==========================================
	// 2. XỬ LÝ PRODUCTS & BƠM CẢ ID LẪN SLUG
	// ==========================================
	prodRows, err := f.GetRows("Products")
	if err != nil {
		log.Fatal("Lỗi đọc sheet Products: ", err)
	}

	var products []interface{}
	for i, row := range prodRows {
		if i == 0 || len(row) == 0 || getCell(row, 0) == "" {
			continue // Bỏ qua tiêu đề
		}

		price, _ := strconv.ParseFloat(getCell(row, 2), 64)
		stock, _ := strconv.Atoi(getCell(row, 3))
		sold, _ := strconv.Atoi(getCell(row, 4))
		discountPrice, _ := strconv.ParseFloat(getCell(row, 5), 64)
		discountPercent, _ := strconv.Atoi(getCell(row, 6))

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

		rawAttr := getCell(row, 13)
		var attributes map[string]interface{}
		if rawAttr != "" {
			if err := json.Unmarshal([]byte(rawAttr), &attributes); err != nil {
				log.Printf("⚠️ Dòng %d: Lỗi JSON cột Attributes: %v\n", i+1, err)
			}
		}

		status := getCell(row, 14)
		if status == "" {
			status = "active"
		}

		// --- ĐIỂM CHUYỂN GIAO QUAN TRỌNG ---
		excelCategoryCode := getCell(row, 9)               // Lấy chữ "C001" ở cột J
		realMongoID := categoryIDMap[excelCategoryCode]    // Ra ID "6a3cb5..."
		realSlug := categorySlugMap[excelCategoryCode]     // Ra chữ "smartphones"

		prod := models.Product{
			ID:              primitive.NewObjectID(),
			Name:            getCell(row, 0),
			Description:     getCell(row, 1),
			Price:           price,
			Stock:           stock,
			Sold:            sold,
			DiscountPrice:   discountPrice,
			DiscountPercent: discountPercent,

			CategoryID:   realMongoID, // Bơm ID Mongo
			CategorySlug: realSlug,    // Bơm luôn cả chữ "smartphones"

			VendorID:   getCell(row, 10),
			Brand:      getCell(row, 11),
			Images:     images,
			Attributes: attributes,
			Status:     status,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

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

	fmt.Println("🚀 HOÀN TẤT BƠM DỮ LIỆU BAO GỒM CẢ ID VÀ SLUG!")
}