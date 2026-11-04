package main

import (
	"log/slog"
	"net/http"
	"os"

	"dailymeal/backend/internal/config"
	"dailymeal/backend/internal/database"
	"dailymeal/backend/internal/httpapi"
	"dailymeal/backend/internal/service"
	"dailymeal/backend/internal/store"
	"dailymeal/backend/openapi"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	repository := store.New(db)
	svc := service.New(repository)
	router := httpapi.New(svc, repository.Ping, openapi.Document)
	slog.Info("listening", "address", cfg.HTTPAddr, "business_timezone", config.Location.String())
	// TODO：退出时暂不等待在途请求完成，请求可能被中断。
	return http.ListenAndServe(cfg.HTTPAddr, router)
}
