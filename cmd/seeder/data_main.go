package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 1. Đọc file .env
	err := godotenv.Load("../../.env")
	if err != nil {
		err = godotenv.Load()
	}

	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "ecommerce_product"
	}

	// 2. Kết nối MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Lỗi kết nối:", err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(dbName).Collection("products")

	// Xóa dữ liệu cũ để tránh rác (Tùy chọn: Nếu muốn giữ thì comment dòng này lại)
	collection.Drop(ctx)
	fmt.Println("Đã dọn dẹp dữ liệu cũ...")

	// 3. Chuẩn bị tập dữ liệu ngẫu nhiên
	categories := []string{"smartphone", "laptop", "accessories", "smartwatch"}
	brands := []string{"Apple", "Samsung", "Xiaomi", "Asus", "Logitech"}
	adjectives := []string{"Pro", "Ultra", "Max", "Gaming", "Chính hãng VN/A"}
	items := []string{"iPhone 15", "Galaxy S24", "MacBook Pro", "Tai nghe AirPods", "Cáp sạc 20W"}

	var products []interface{}
	rand.Seed(time.Now().UnixNano())

	fmt.Println("⏳ Đang tiến hành sinh dữ liệu hàng loạt...")

	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("%s %s %s", items[rand.Intn(len(items))], brands[rand.Intn(len(brands))], adjectives[rand.Intn(len(adjectives))])
		
		originalPrice := float64(rand.Intn(500)+10) * 10000 // Giá từ 100k đến 5 triệu
		
		// Khởi tạo các biến Flash Sale mặc định (Không sale)
		var discountPrice float64 = 0
		var discountPercent int = 0
		var saleStartDate *time.Time = nil
		var saleEndDate *time.Time = nil

		// Logic tạo Flash Sale: 20% cơ hội sản phẩm được Sale
		if rand.Intn(100) < 20 { 
			// Chọn ngẫu nhiên % giảm giá: 10, 20, 30, 40 hoặc 50%
			discountPercents := []int{10, 20, 30, 40, 50}
			discountPercent = discountPercents[rand.Intn(len(discountPercents))]
			
			// Tính giá sau khi giảm
			discountPrice = originalPrice - (originalPrice * float64(discountPercent) / 100)
			
			// Thiết lập thời gian Sale: Bắt đầu từ hiện tại, kết thúc trong 24h-48h tới
			now := time.Now()
			end := now.Add(time.Duration(rand.Intn(24) + 24) * time.Hour)
			
			saleStartDate = &now
			saleEndDate = &end
		}

		product := bson.M{
			"_id":              primitive.NewObjectID(),
			"name":             name,
			"description":      fmt.Sprintf("Mô tả chi tiết cho sản phẩm %s. Đảm bảo chất lượng tuyệt đối.", name),
			"price":            originalPrice,
			"stock":            rand.Intn(200) + 10, // Số lượng trong kho
			"sold":             rand.Intn(150),      // Số lượng đã bán
			"discount_price":   discountPrice,
			"discount_percent": discountPercent,
			"sale_start_date":  saleStartDate,
			"sale_end_date":    saleEndDate,
			"category_id":      categories[rand.Intn(len(categories))],
			"vendor_id":        fmt.Sprintf("vendor_%d", rand.Intn(5)+1),
			"brand":            brands[rand.Intn(len(brands))],
			"images":           []string{},
			"attributes": bson.M{
				"warranty": "12 tháng",
				"color":  "Đen/Trắng",
			},
			"status":     "ACTIVE",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		products = append(products, product)
	}

	// 4. Insert vào MongoDB
	_, err = collection.InsertMany(context.Background(), products)
	if err != nil {
		log.Fatal("Lỗi khi chèn dữ liệu:", err)
	}

	fmt.Printf("✅ Đã thêm thành công %d sản phẩm vào MongoDB Atlas (Bao gồm các sản phẩm Flash Sale)!\n", len(products))
}