package config

import (
	"fmt"
	"os"
	"time"
	_ "time/tzdata"
)

type Config struct {
	DatabaseURL string
	HTTPAddr    string
	Location    *time.Location
}

func Load() (Config, error) {
	cfg := Config{DatabaseURL: os.Getenv("DATABASE_URL"), HTTPAddr: os.Getenv("HTTP_ADDR")}
	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	zone := os.Getenv("BUSINESS_TIMEZONE")
	if zone == "" {
		zone = "Asia/Shanghai"
	}
	var err error
	cfg.Location, err = time.LoadLocation(zone)
	if err != nil {
		return cfg, fmt.Errorf("invalid BUSINESS_TIMEZONE: %w", err)
	}
	return cfg, nil
}
