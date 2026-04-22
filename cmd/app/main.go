package main

import (
	"log"
	"net/http"

	"effective-go/internal/config"
	"effective-go/internal/handler"
	"effective-go/internal/repository"
	"effective-go/internal/service"

	_ "effective-go/docs"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Subscription API
// @version 1.0
// @description REST-сервис для агрегации данных об онлайн-подписках пользователей.
// @host localhost:8080
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg := config.Load()

	db, err := config.NewDB(cfg)
	if err != nil {
		log.Fatal("db connect error: ", err)
	}
	defer db.Close()

	subRepo := repository.NewSubscriptionRepo(db)
	subService := service.NewSubscriptionService(subRepo)
	subHandler := handler.NewSubscriptionHandler(subService)

	r := chi.NewRouter()

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

	log.Printf("server started on :%s", cfg.AppPort)
	if err = http.ListenAndServe(":"+cfg.AppPort, r); err != nil {
		log.Fatal(err)
	}
}
