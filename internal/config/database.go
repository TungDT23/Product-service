package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database
var RedisClient *redis.Client

func ConnectDB() {
	// Kết nối MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://root:secret@localhost:27017")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Lỗi kết nối MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Không thể ping MongoDB:", err)
	}

	fmt.Println("Đã kết nối thành công tới MongoDB!")
	DB = client.Database("ecommerce_product")
}

func ConnectRedis() {
	// Kết nối Redis
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Lỗi kết nối Redis:", err)
	}

	fmt.Println("Đã kết nối thành công tới Redis!")
}
