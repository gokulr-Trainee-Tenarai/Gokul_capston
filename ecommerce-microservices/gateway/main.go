package main

import (
	"io"
	"log"
	"net/http"
	"strings"

	"ecommerce-microservices/gateway/client"
	"ecommerce-microservices/gateway/config"
	gqlserver "ecommerce-microservices/gateway/graphql"
	"ecommerce-microservices/pkg/auth"
	"ecommerce-microservices/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func createDynamicProxy(consulClient *client.ConsulClient, serviceName, prefixToRemove string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetBaseURL, err := consulClient.GetServiceURL(serviceName)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Consul discovery failed: " + err.Error()})
			return
		}

		targetPath := strings.TrimPrefix(c.Request.URL.Path, prefixToRemove)
		if targetPath == "" {
			targetPath = "/"
		}

		targetURL := targetBaseURL + targetPath
		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		req.Header = c.Request.Header.Clone()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "downstream service unavailable"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

func main() {
	cfg := config.Load()
	consulClient := client.NewConsulClient(cfg.ConsulURL)

	gs, err := gqlserver.NewServer(consulClient)
	if err != nil {
		log.Fatalf("GraphQL schema failed: %v", err)
	}

	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware("api-gateway"))
	r.Use(middleware.StructuredLogger("api-gateway"))

	// In-memory valid credentials and roles
	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	validUsers := map[string]struct {
		Password string
		Role     string
	}{
		"admin": {Password: "admin123", Role: "ADMIN"},
		"user":  {Password: "user123", Role: "USER"},
	}

	// 1. POST Login Endpoint
	r.POST("/api/v1/auth/token", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required in JSON body"})
			return
		}

		user, exists := validUsers[strings.ToLower(strings.TrimSpace(req.Username))]
		if !exists || user.Password != req.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		token, err := auth.GenerateToken(req.Username, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"username":     req.Username,
			"role":         user.Role,
			"token":        token,
			"header_value": "Bearer " + token,
			"instructions": "Copy header_value and paste into SoapUI Authorization header",
		})
	})

	// Quick GET fallback endpoints
	r.GET("/api/v1/auth/token/user", func(c *gin.Context) {
		token, err := auth.GenerateToken("test-user", "USER")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"role":         "USER",
			"token":        token,
			"header_value": "Bearer " + token,
		})
	})

	r.GET("/api/v1/auth/token/admin", func(c *gin.Context) {
		token, err := auth.GenerateToken("test-admin", "ADMIN")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"role":         "ADMIN",
			"token":        token,
			"header_value": "Bearer " + token,
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "api-gateway", "status": "UP"})
	})

	// GraphQL
	h := gqlserver.NewHTTPHandler(gs)
	r.GET("/graphql", gin.WrapH(h))
	r.POST("/graphql", gin.WrapH(h)) // Open for browser Playground
	r.Any("/playground", gin.WrapH(h))

	// Products routes (Supports both /api/products and /api/v1/products)
	r.GET("/api/products", createDynamicProxy(consulClient, "product-service", "/api"))
	r.GET("/api/products/*path", createDynamicProxy(consulClient, "product-service", "/api"))
	r.GET("/api/v1/products", createDynamicProxy(consulClient, "product-service", "/api/v1"))
	r.GET("/api/v1/products/*path", createDynamicProxy(consulClient, "product-service", "/api/v1"))
	r.POST("/api/products", auth.JWTMiddleware("ADMIN"), createDynamicProxy(consulClient, "product-service", "/api"))
	r.POST("/api/v1/products", auth.JWTMiddleware("ADMIN"), createDynamicProxy(consulClient, "product-service", "/api/v1"))

	// Orders routes
	orderUserGroup := r.Group("/api/v1/orders")
	orderUserGroup.Use(auth.JWTMiddleware("USER", "ADMIN"))
	{
		orderUserGroup.POST("", createDynamicProxy(consulClient, "order-service", "/api/v1"))
		orderUserGroup.GET("", createDynamicProxy(consulClient, "order-service", "/api/v1"))
		orderUserGroup.GET("/:id", createDynamicProxy(consulClient, "order-service", "/api/v1"))
		orderUserGroup.GET("/:id/product", createDynamicProxy(consulClient, "order-service", "/api/v1"))
	}

	orderAdminGroup := r.Group("/api/v1/orders")
	orderAdminGroup.Use(auth.JWTMiddleware("ADMIN"))
	{
		orderAdminGroup.PUT("/:id/status", createDynamicProxy(consulClient, "order-service", "/api/v1"))
	}

	log.Println("API Gateway running at http://localhost:" + cfg.Port)
	_ = r.Run(":" + cfg.Port)
}
