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
	// 1. Đọc file .env (lùi lại 2 cấp thư mục để tìm file .env ở gốc)
	err := godotenv.Load("../../.env")
	if err != nil {
		err = godotenv.Load() // Fallback thử tìm ở thư mục hiện tại
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

	// 3. Chuẩn bị tập dữ liệu ngẫu nhiên
	categories := []string{"cat_electronics", "cat_fashion", "cat_home", "cat_books"}
	brands := []string{"Samsung", "Apple", "Sony", "Nike", "Adidas", "LG"}
	adjectives := []string{"Cao cấp", "Thông minh", "Thế hệ mới", "Giá rẻ", "Chính hãng"}
	items := []string{"Điện thoại", "Tai nghe", "Áo khoác", "Balo", "Tủ lạnh", "Sách lập trình"}

	var products []interface{}
	rand.Seed(time.Now().UnixNano())

	fmt.Println("⏳ Đang tiến hành sinh dữ liệu hàng loạt...")

	// Vòng lặp sinh 100 sản phẩm (Bạn có thể đổi số 100 thành 1000 hoặc 5000 tùy ý)
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("%s %s %s", items[rand.Intn(len(items))], brands[rand.Intn(len(brands))], adjectives[rand.Intn(len(adjectives))])
		
		product := bson.M{
			"_id":         primitive.NewObjectID(),
			"name":        name,
			"description": fmt.Sprintf("Mô tả chi tiết cho sản phẩm %s. Chất lượng đảm bảo 100%%.", name),
			"price":       float64(rand.Intn(500)+10) * 10000, // Giá từ 100k đến 5 triệu
			"stock":       rand.Intn(200) + 10,
			"category_id": categories[rand.Intn(len(categories))],
			"vendor_id":   fmt.Sprintf("vendor_%d", rand.Intn(5)+1),
			"brand":       brands[rand.Intn(len(brands))],
			"images":      []string{},
			"attributes": bson.M{
				"weight": fmt.Sprintf("%d kg", rand.Intn(5)+1),
				"color":  "Đen/Trắng",
			},
			"status":     "ACTIVE",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
		products = append(products, product)
	}

	// 4. Insert 1 cục vào MongoDB (InsertMany để tối ưu tốc độ)
	_, err = collection.InsertMany(context.Background(), products)
	if err != nil {
		log.Fatal("Lỗi khi chèn dữ liệu:", err)
	}

	fmt.Printf("✅ Đã thêm thành công %d sản phẩm vào MongoDB Atlas!\n", len(products))
}