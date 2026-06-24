package middlewares

import (
	"fmt"
	"net/http"
	"os"
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

		// 💡 Tiền hành giải mã Token bằng JWT_SECRET
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Kiểm tra xem thuật toán mã hóa có đúng chuẩn không
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("thuật toán mã hóa không hợp lệ: %v", token.Header["alg"])
			}
			// Lấy chìa khóa bí mật từ file .env ra để mở khóa
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ hoặc đã hết hạn"})
			c.Abort()
			return
		}

		// 💡 Lấy thông tin (Claims) từ trong Token ra
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Lấy trường role (quyền) ra. 
			// Lưu ý: Bạn cần hỏi lại bạn kia xem họ đặt tên trường quyền là "role", "is_admin" hay gì khác nhé!
			role, ok := claims["role"].(string)
			if !ok {
				role = "user" // Nếu không có trường role, mặc định coi là user bình thường
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

// Kiểm tra quyền Admin (Giữ nguyên như cũ)
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