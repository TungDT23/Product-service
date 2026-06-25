package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDB() {
	// 1. Load các biến môi trường từ file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Cảnh báo: Không tìm thấy file .env, hệ thống sẽ sử dụng biến môi trường của OS")
	}

	// 2. Lấy thông tin từ file .env
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	if mongoURI == "" {
		log.Fatal("Lỗi: Chưa cấu hình biến môi trường MONGO_URI trong file .env")
	}
	if dbName == "" {
		dbName = "ecommerce_product_db" // Giữ nguyên tên DB cũ của bạn
	}

	// 3. Kết nối MongoDB Cloud
	clientOptions := options.Client().ApplyURI(mongoURI)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Lỗi kết nối MongoDB:", err)
	}

	// 4. Ping kiểm tra kết nối
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Không thể ping MongoDB:", err)
	}

	fmt.Println("Đã kết nối thành công tới MongoDB Atlas!")
	DB = client.Database(dbName)
}

