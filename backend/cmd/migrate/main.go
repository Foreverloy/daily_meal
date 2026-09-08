package main

import (
	"fmt"
	"log/slog"
	"os"

	"dailymeal/backend/internal/config"
	"dailymeal/backend/internal/database"
	"dailymeal/backend/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command != "up" && command != "down" && command != "status" {
		return fmt.Errorf("usage: migrate [up|down|status]")
	}
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
	return migrations.Run(sqlDB, command)
}
