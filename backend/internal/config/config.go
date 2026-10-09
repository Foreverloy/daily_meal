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
}

// Location is loaded from BUSINESS_TIMEZONE at startup; tzdata includes the default zone.
var Location, _ = time.LoadLocation("Asia/Shanghai")

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
	location, err := time.LoadLocation(zone)
	if err != nil {
		return cfg, fmt.Errorf("invalid BUSINESS_TIMEZONE: %w", err)
	}
	Location = location
	return cfg, nil
}
