package config

import (
	"fmt"
	"os"
	"time"
	_ "time/tzdata"
)

// Config 保存 Go 服务启动时从环境变量读取的数据库和 HTTP 配置。
type Config struct {
	// DatabaseURL 是 PostgreSQL 连接串，来自 DATABASE_URL，如 postgres://localhost/daily_meal。
	DatabaseURL string
	// HTTPAddr 是监听地址，来自 HTTP_ADDR；默认 :8080 表示监听本机的 8080 端口。
	HTTPAddr string
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
