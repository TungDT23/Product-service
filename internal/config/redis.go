package config

import (
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

var RedisClient *redis.Client

func ConnectRedis() {
	// Lấy đường link Redis từ file .env hoặc Render
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("Lỗi: Không tìm thấy biến môi trường REDIS_URL")
	}

	// Phân tích đường link tự động
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Lỗi định dạng cấu hình Redis:", err)
	}

	// Khởi tạo client
	RedisClient = redis.NewClient(opt)

	// Ping thử xem có thông mạng không
	_, err = RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Println("Cảnh báo: Không thể kết nối Redis Cloud:", err)
	} else {
		log.Println("Đã kết nối thành công tới Redis Cloud (Upstash)!")
	}
}