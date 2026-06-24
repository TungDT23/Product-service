package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Giả lập hàm giải mã Token (Thực tế bạn sẽ dùng thư viện jwt-go để đọc token từ Auth Service)
func parseTokenAndGetRole(token string) string {
	// MOCK LOGIC: Để test tạm, ta tự quy định:
	// - Nếu token là "token_cua_admin" -> Role = admin
	// - Nếu token là "token_cua_user" -> Role = user
	if token == "token_cua_admin" {
		return "admin"
	} else if token == "token_cua_user" {
		return "user"
	}
	return ""
}

// Kiểm tra xem có đăng nhập chưa (Ai cũng được, miễn có token)
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Thiếu token đăng nhập"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		role := parseTokenAndGetRole(token)
		
		if role == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không hợp lệ"})
			c.Abort()
			return
		}

		// Lưu role vào context để các hàm sau dùng
		c.Set("role", role)
		c.Next()
	}
}

// Kiểm tra quyền Admin (Chỉ dành cho Admin)
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