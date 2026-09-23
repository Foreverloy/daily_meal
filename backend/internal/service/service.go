package service

import (
	"errors"
	"strings"
	"time"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/store"
	"gorm.io/gorm"
)

type Service struct {
	store    *store.Store
	location *time.Location
	now      func() time.Time
}

func New(db *store.Store, location *time.Location, now func() time.Time) *Service {
	return &Service{store: db, location: location, now: now}
}

func (s *Service) Today() model.Date { return model.Date(s.now().In(s.location).Format(time.DateOnly)) }

func ParseDate(date string) (model.Date, error) {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil || t.Format(time.DateOnly) != date || t.Year() < 1 {
		return "", Invalid("date", "日期必须为有效的 YYYY-MM-DD")
	}
	return model.Date(date), nil
}

func cleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 200 {
		return "", Invalid("name", "名称须为 1～200 个字符")
	}
	return name, nil
}

func resourceError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NotFound()
	}
	return err
}

// All multi-query reads use one PostgreSQL snapshot, including recipe food preloads.
func (s *Service) transaction(fn func(*Service) error) error {
	return s.store.Transaction(func(tx *store.Store) error {
		return fn(&Service{store: tx, location: s.location, now: s.now})
	})
}

type Page struct {
	Query  string
	Number int
	Size   int
}

func (p Page) Validate() error {
	if p.Number < 1 || p.Number > 1000000 {
		return Invalid("page", "页码须为 1～1000000")
	}
	if p.Size < 1 || p.Size > 100 {
		return Invalid("page_size", "每页须为 1～100 条")
	}
	return nil
}

type List[T any] struct {
	Items    []T   `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}
