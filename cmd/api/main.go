package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gsupert/internal/common"
	"gsupert/internal/config"
	"gsupert/internal/db"
	"gsupert/internal/modules/auth"
	"gsupert/internal/modules/customers"
	"gsupert/internal/modules/notes"
	"gsupert/internal/modules/settings"
	textanalyze "gsupert/internal/modules/text_analyze"
	topicconversation "gsupert/internal/modules/topic_conversation"
	"gsupert/internal/modules/users"
)

func main() {
	cfg := config.LoadConfig()
	if err := db.RunMigrations(cfg); err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}
	database := db.InitDB(cfg)
	emailService := common.NewEmailService(cfg)

	r := gin.Default()

	// Error Logger Middleware (Logs only errors >= 400)
	r.Use(common.ErrorLogger())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize Repositories
	userRepo := users.NewRepository(database)
	customerRepo := customers.NewRepository(database)
	noteRepo := notes.NewRepository(database)
	settingsRepo := settings.NewRepository(database)

	// Initialize Services
	userService := users.NewService(userRepo, cfg)
	customerService := customers.NewService(customerRepo, emailService)
	noteService := notes.NewService(noteRepo, emailService)
	settingsService := settings.NewService(settingsRepo)

	envOpenAIAPIKey := strings.TrimSpace(cfg.OpenAIAPIKey)
	openAIAPIKey := envOpenAIAPIKey
	if openAIAPIKey == "" {
		setting, err := settingsService.GetByKey(string(settings.SettingKeyGPTAPIKey))
		if err == nil {
			openAIAPIKey = strings.TrimSpace(setting.Value)
			if openAIAPIKey != "" {
				log.Printf("OpenAI API key loaded from app setting '%s'", settings.SettingKeyGPTAPIKey)
			}
		} else if !errors.Is(err, settings.ErrSettingNotFound) {
			log.Printf("Failed to load OpenAI API key from app settings: %v", err)
		}
	}

	defaultSyntaxProvider := textanalyze.NewMockSyntaxProvider()
	gptSyntaxProvider := textanalyze.NewOpenAISyntaxProvider(
		openAIAPIKey,
		cfg.OpenAIModel,
		cfg.OpenAIBaseURL,
		nil,
	)
	if gptSyntaxProvider != nil {
		log.Printf("Text analyze GPT syntax provider enabled")
	}
	textAnalyzeService := textanalyze.NewService(defaultSyntaxProvider, gptSyntaxProvider)
	topicConversationProvider := topicconversation.NewOpenAIProvider(
		envOpenAIAPIKey,
		cfg.OpenAIModel,
		cfg.OpenAIBaseURL,
		nil,
		func(ctx context.Context) (string, error) {
			setting, err := settingsService.GetByKey(string(settings.SettingKeyGPTAPIKey))
			if err != nil {
				if errors.Is(err, settings.ErrSettingNotFound) {
					return "", nil
				}
				return "", err
			}
			return strings.TrimSpace(setting.Value), nil
		},
	)
	if topicConversationProvider != nil {
		log.Printf("Topic conversation GPT provider enabled")
	}
	topicConversationService := topicconversation.NewService(topicConversationProvider)

	// Initialize Handlers
	userHandler := users.NewHandler(userService)
	customerHandler := customers.NewHandler(customerService)
	noteHandler := notes.NewHandler(noteService)
	settingsHandler := settings.NewHandler(settingsService)
	textAnalyzeHandler := textanalyze.NewHandler(textAnalyzeService)
	topicConversationHandler := topicconversation.NewHandler(topicConversationService)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Text analyze (MVP)
	r.POST("/text-analyze", textAnalyzeHandler.Analyze)
	r.POST("/topic-conversation", topicConversationHandler.Generate)

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
			customerRoutes.GET("/export/pdf", customerHandler.ExportPDF)
			customerRoutes.GET("/export/excel", customerHandler.ExportExcel)
			customerRoutes.POST("/:id/send-email", customerHandler.SendEmail)
			customerRoutes.GET("/:id", customerHandler.GetCustomer)
			customerRoutes.POST("", customerHandler.CreateCustomer)
			customerRoutes.PUT("/:id", customerHandler.UpdateCustomer)
			customerRoutes.DELETE("/:id", customerHandler.DeleteCustomer)
		}

		// Notes CRUD
		noteRoutes := api.Group("/notes")
		{
			noteRoutes.GET("", noteHandler.ListNotes)
			noteRoutes.GET("/:id", noteHandler.GetNote)
			noteRoutes.POST("", noteHandler.CreateNote)
			noteRoutes.PUT("/:id", noteHandler.UpdateNote)
			noteRoutes.DELETE("/:id", noteHandler.DeleteNote)
		}

		// Settings (Admin only)
		settingRoutes := api.Group("/settings")
		settingRoutes.Use(auth.RoleMiddleware("admin"))
		{
			settingRoutes.POST("/bulk", settingsHandler.UpsertSettings)
			settingRoutes.GET("", settingsHandler.ListSettings)
			settingRoutes.GET("/:key", settingsHandler.GetSettingByKey)
		}
	}

	log.Printf("Server starting on port %s", cfg.AppPort)
	if err := r.Run(fmt.Sprintf(":%s", cfg.AppPort)); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}
