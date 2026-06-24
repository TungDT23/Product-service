package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Kiểm tra xem có đăng nhập chưa (Ai cũng được, miễn có token hợp lệ)
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu token đăng nhập"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Tiền hành giải mã Token bằng JWT_SECRET
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Kiểm tra xem thuật toán mã hóa có đúng chuẩn không
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("thuật toán mã hóa không hợp lệ: %v", token.Header["alg"])
			}
			// Vẫn giữ chuỗi hardcode để test lỗi
			return []byte("lkasfjdslkdfjlasfjl"), nil
		})

		// 💡 ĐOẠN ĐÃ SỬA: Bắt và in lỗi chi tiết ra Postman
		if err != nil {
			fmt.Println("LỖI GIẢI MÃ TOKEN:", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":        "Quá trình giải mã thất bại",
				"chi_tiet_loi": err.Error(),
			})
			c.Abort()
			return
		}

		// 💡 ĐOẠN ĐÃ SỬA: Tách riêng phần kiểm tra tính hợp lệ
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token đã hết hạn hoặc bị hỏng"})
			c.Abort()
			return
		}

		// Lấy thông tin (Claims) từ trong Token ra
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			role, ok := claims["role"].(string)
			if !ok {
				role = "user"
			}
			
			// Lưu role vào context để hàm RequireAdmin phía sau có thể kiểm tra
			c.Set("role", role)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Không đọc được dữ liệu token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Kiểm tra quyền Admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Chỉ Admin mới có quyền thực hiện hành động này"})
			c.Abort()
			return
		}
		c.Next()
	}
}