package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/provemyself/backend/internal/config"
	"github.com/provemyself/backend/internal/core"
	"github.com/provemyself/backend/internal/http/handlers"
	"github.com/provemyself/backend/internal/middleware"
	"github.com/provemyself/backend/internal/store"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Caller().
		Logger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Initialize validator
	validate := validator.New()

	// Initialize database
	database, err := store.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize database")
	}
	defer database.Close()

	// Run database migrations
	ctx := context.Background()
	if err := database.Migrate(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to run database migrations")
	}

	// Initialize stores
	projectStore := store.NewProjectStore(database)
	itemStore := store.NewItemStore(database)

	// Initialize services
	projectService := core.NewProjectService(projectStore)
	itemService := core.NewItemService(itemStore, projectStore)

	// Initialize middleware
	loggingMiddleware := middleware.NewLoggingMiddleware()
	healthMiddleware := middleware.NewHealthMiddleware()
	errorHandler := middleware.NewErrorHandler()
	validationMiddleware := middleware.NewValidationMiddleware()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(database)
	projectHandler := handlers.NewProjectHandler(projectService, validate)
	itemHandler := handlers.NewItemHandler(itemService, validate)

	// Setup router
	r := chi.NewRouter()

	// Core middleware stack
	r.Use(loggingMiddleware.RequestID)
	r.Use(loggingMiddleware.UserContext)
	r.Use(loggingMiddleware.RequestLogger)
	r.Use(errorHandler.Recovery)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health and monitoring endpoints (outside API versioning)
	r.Get("/health", healthHandler.GetHealth)
	r.Get("/health/live", healthMiddleware.LivenessProbe)
	r.Get("/health/ready", healthMiddleware.ReadinessProbe([]middleware.HealthChecker{
		middleware.NewDatabaseHealthChecker("database", database.HealthCheck),
	}))
	r.Get("/metrics", healthMiddleware.Metrics)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Projects
		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projectHandler.ListProjects)
			r.Post("/", projectHandler.CreateProject)
			r.Get("/{projectId}", projectHandler.GetProject)
			r.Put("/{projectId}", projectHandler.UpdateProject)
			r.Delete("/{projectId}", projectHandler.DeleteProject)
			r.Post("/{projectId}/publish", projectHandler.PublishProject)

			// Items nested under projects
			r.Route("/{projectId}/items", func(r chi.Router) {
				r.Get("/", itemHandler.ListItems)
				r.Post("/", itemHandler.CreateItem)
				r.Get("/{itemId}", itemHandler.GetItem)
				r.Put("/{itemId}", itemHandler.UpdateItem)
				r.Delete("/{itemId}", itemHandler.DeleteItem)
				
				// Bulk operations and position management
				r.Post("/bulk", itemHandler.BulkCreateItems)
				r.Put("/positions", itemHandler.UpdateItemPositions)
			})
		})
	})

	// Server configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info().
			Str("addr", srv.Addr).
			Msg("starting server")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server forced to shutdown")
	}

	logger.Info().Msg("server exited")
}