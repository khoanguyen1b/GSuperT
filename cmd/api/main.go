package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gsupert/internal/config"
	"gsupert/internal/db"
	"gsupert/internal/modules/auth"
	"gsupert/internal/modules/customers"
	"gsupert/internal/modules/users"
)

func main() {
	cfg := config.LoadConfig()
	database := db.InitDB(cfg)

	r := gin.Default()

	// Initialize Repositories
	userRepo := users.NewRepository(database)
	customerRepo := customers.NewRepository(database)

	// Initialize Services
	userService := users.NewService(userRepo, cfg)
	customerService := customers.NewService(customerRepo)

	// Initialize Handlers
	userHandler := users.NewHandler(userService)
	customerHandler := customers.NewHandler(customerService)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", userHandler.Login)
		authGroup.POST("/refresh", userHandler.RefreshToken)
		authGroup.POST("/logout", auth.AuthMiddleware(cfg), userHandler.Logout)
	}

	// Protected routes
	api := r.Group("/")
	api.Use(auth.AuthMiddleware(cfg))
	{
		// Users CRUD (Admin only)
		userRoutes := api.Group("/users")
		userRoutes.Use(auth.RoleMiddleware("admin"))
		{
			userRoutes.GET("", userHandler.ListUsers)
			userRoutes.GET("/:id", userHandler.GetUser)
			userRoutes.POST("", userHandler.CreateUser)
			userRoutes.PUT("/:id", userHandler.UpdateUser)
			userRoutes.DELETE("/:id", userHandler.DeleteUser)
		}

		// Customers CRUD
		customerRoutes := api.Group("/customers")
		{
			customerRoutes.GET("", customerHandler.ListCustomers)
			customerRoutes.GET("/:id", customerHandler.GetCustomer)
			customerRoutes.POST("", customerHandler.CreateCustomer)
			customerRoutes.PUT("/:id", customerHandler.UpdateCustomer)
			customerRoutes.DELETE("/:id", customerHandler.DeleteCustomer)
		}
	}

	log.Printf("Server starting on port %s", cfg.AppPort)
	if err := r.Run(fmt.Sprintf(":%s", cfg.AppPort)); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
