// Package store contains database queries. Transactions are chosen by services.
package store

import (
	"context"
	"database/sql"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct{ db *gorm.DB }

func New(db *gorm.DB) *Store { return &Store{db: db} }

func (s *Store) Transaction(fn func(*Store) error) error {
	return s.db.Transaction(func(tx *gorm.DB) error { return fn(New(tx)) }, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
}

func (s *Store) Ping(ctx context.Context) error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.PingContext(ctx)
}

func lock(query *gorm.DB, enabled bool) *gorm.DB {
	if enabled {
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	return query
}

func named(query *gorm.DB, q string) *gorm.DB {
	if q == "" {
		return query
	}
	q = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
	return query.Where("name ILIKE ?", "%"+q+"%")
}
