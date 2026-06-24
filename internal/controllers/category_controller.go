package controllers

import (
	"net/http"
	"product-service/internal/services"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	Service *services.ProductService
}

func NewCategoryController(service *services.ProductService) *CategoryController {
	return &CategoryController{
		Service: service,
	}
}

func (c *CategoryController) GetAllCategories(ctx *gin.Context) {
	categories, err := c.Service.Repo.GetAllCategories(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Không thể lấy danh mục"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}