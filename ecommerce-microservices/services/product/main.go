package main

import (
	"log"
	"net/http"
	"context"

	"ecommerce-microservices/database"
	"ecommerce-microservices/pkg/middleware"
	"ecommerce-microservices/services/product/config"
	"ecommerce-microservices/services/product/controller"
	"ecommerce-microservices/services/product/model"
	"ecommerce-microservices/services/product/repository"
	"ecommerce-microservices/services/product/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// 1. Initialize GORM connection pool
	poolCfg := database.DefaultPoolConfig(cfg.DSN())
	db, err := database.Connect(context.Background(), poolCfg) // or database.Connect(context.Background(), poolCfg) depending on your signature
	if err != nil {
		log.Fatalf("product database connection error: %v", err)
	}

	// 2. Run GORM auto-migration for Product model
	if err := db.AutoMigrate(&model.Product{}); err != nil {
		log.Fatalf("product schema migration failed: %v", err)
	}

	// 3. Extract the underlying *sql.DB for health checks
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to retrieve sql.DB from gorm: %v", err)
	}

	repo := repository.NewPostgresProductRepository(db)
	svc := service.NewProductService(repo)
	ctrl := controller.NewProductController(svc)

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware("product-service"))
	r.Use(middleware.StructuredLogger("product-service"))

	// Liveness & Basic Probes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "product-service", "status": "UP"})
	})
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "product-service", "status": "ALIVE"})
	})

	// Readiness Probe: Uses sqlDB.PingContext to verify DB connection
	r.GET("/health/ready", func(c *gin.Context) {
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "NOT_READY",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})

	// Business Routes
	r.POST("/products", ctrl.Create)
	r.GET("/products/:id", ctrl.GetByID)
	r.GET("/products", ctrl.GetAll)

	log.Println("Product Service listening on :" + cfg.Port)
	_ = r.Run(":" + cfg.Port)
}