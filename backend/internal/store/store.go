// Package store contains database queries.
package store

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

// Store 封装 PostgreSQL 数据读写，供 Service 调用。
type Store struct {
	// db 是 GORM 数据库句柄，通过它执行查询、插入、更新和删除。
	db *gorm.DB
}

func New(db *gorm.DB) *Store { return &Store{db: db} }

func (s *Store) Ping(ctx context.Context) error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

func named(query *gorm.DB, q string) *gorm.DB {
	if q == "" {
		return query
	}
	q = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
	return query.Where("name ILIKE ?", "%"+q+"%")
}
