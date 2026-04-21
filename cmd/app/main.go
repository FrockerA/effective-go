package main

import (
	"log"
	"net/http"

	"effective-go/internal/config"
	"effective-go/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg := config.Load()

	db, err := config.NewDB(cfg)
	if err != nil {
		log.Fatal("db connect error: ", err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Get("/health", handler.Health)

	log.Printf("server started on :%s", cfg.AppPort)
	if err = http.ListenAndServe(":"+cfg.AppPort, r); err != nil {
		log.Fatal(err)
	}
}
