package controller

import (
	"net/http"
	"strconv"

	"ecommerce-microservices/services/product/model"
	"ecommerce-microservices/services/product/repository"
	"ecommerce-microservices/services/product/service"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	service service.ProductService
}

func NewProductController(s service.ProductService) *ProductController {
	return &ProductController{service: s}
}

func (c *ProductController) Create(ctx *gin.Context) {
	var p model.Product
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON request body"})
		return
	}
	if err := c.service.Create(ctx.Request.Context(), &p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, p)
}

func (c *ProductController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "product id must be numeric"})
		return
	}
	p, err := c.service.GetByID(ctx.Request.Context(), id)
	if repository.IsNotFound(err) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, p)
}

func (c *ProductController) GetAll(ctx *gin.Context) {
	products, err := c.service.GetAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, products)
}
