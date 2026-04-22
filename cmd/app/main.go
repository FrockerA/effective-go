package main

import (
	"net/http"
	"time"

	"effective-go/internal/config"
	"effective-go/internal/handler"
	"effective-go/internal/repository"
	"effective-go/internal/service"

	_ "effective-go/docs"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// @title Subscription API
// @version 1.0
// @description REST-сервис для агрегации данных об онлайн-подписках пользователей.
// @host localhost:8080
// @BasePath /
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn("no .env file, reading from environment")
	}

	cfg := config.Load()

	db, err := config.NewDB(cfg)
	if err != nil {
		logger.Fatal("db connect error", zap.Error(err))
	}
	defer db.Close()

	subRepo := repository.NewSubscriptionRepo(db)
	subService := service.NewSubscriptionService(subRepo)

	subHandler := handler.NewSubscriptionHandler(subService, logger)

	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("incoming request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	})

	r.Get("/health", handler.Health)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", subHandler.Create)
		r.Get("/", subHandler.List)
		r.Get("/total", subHandler.CalculateTotal)

		r.Get("/{id}", subHandler.GetByID)
		r.Put("/{id}", subHandler.Update)
		r.Delete("/{id}", subHandler.Delete)
	})

	logger.Info("server started", zap.String("port", cfg.AppPort))
	if err = http.ListenAndServe(":"+cfg.AppPort, r); err != nil {
		logger.Fatal("server crashed", zap.Error(err))
	}
}
