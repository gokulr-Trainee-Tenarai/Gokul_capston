package main

import (
	"context"
	"log"
	"net/http"

	"ecommerce-microservices/database"
	"ecommerce-microservices/pkg/middleware"
	"ecommerce-microservices/services/order/config"
	"ecommerce-microservices/services/order/controller"
	"ecommerce-microservices/services/order/model"
	"ecommerce-microservices/services/order/repository"
	"ecommerce-microservices/services/order/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// 1. Initialize GORM connection pool with context
	poolCfg := database.DefaultPoolConfig(cfg.DSN())
	db, err := database.Connect(context.Background(), poolCfg)
	if err != nil {
		log.Fatalf("order database connection error: %v", err)
	}

	// 2. Run GORM auto-migration for Order model
	if err := db.AutoMigrate(&model.Order{}); err != nil {
		log.Fatalf("order schema migration failed: %v", err)
	}

	// 3. Extract sql.DB for the readiness ping probe
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to retrieve sql.DB from gorm: %v", err)
	}

	repo := repository.NewPostgresOrderRepository(db)
	svc := service.NewOrderService(repo, cfg.ProductServiceURL)
	ctrl := controller.NewOrderController(svc)

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware("order-service"))
	r.Use(middleware.StructuredLogger("order-service"))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "order-service", "status": "UP"})
	})
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "order-service", "status": "ALIVE"})
	})
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

	r.POST("/orders", ctrl.Create)
	r.GET("/orders", ctrl.ListAll)
	r.GET("/orders/:id", ctrl.GetByID)
	r.PUT("/orders/:id/status", ctrl.UpdateStatus)
	r.GET("/orders/:id/product", ctrl.GetOrderWithProduct)

	log.Println("Order Service listening on :" + cfg.Port)
	_ = r.Run(":" + cfg.Port)
}