package service

import (
	"time"

	"dailymeal/backend/internal/config"
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/store"
)

// Service 承接 Handler 的业务请求，处理营养计算、快照和参数转换，再调用 Store 保存数据。
type Service struct {
	// store 是数据访问层；业务处理通过它读取和写入 PostgreSQL。
	store *store.Store
}

func New(db *store.Store) *Service {
	return &Service{store: db}
}

func (s *Service) Today() model.Date {
	return model.Date(time.Now().In(config.Location).Format(time.DateOnly))
}

func ParseDate(date string) (model.Date, error) {
	_, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return "", err
	}
	return model.Date(date), nil
}

// Page 保存列表查询条件，由 HTTP 查询参数转换而来，食物和菜品列表共用。
type Page struct {
	// Favorite 仅用于食物查询：true 只查已收藏，nil 或 false 查询全部。
	Favorite *bool
	// Query 是名称搜索词，对应 URL 中的 q，例如 q=牛奶。
	Query string
	// Number 是从 1 开始的页码，对应 page；接口默认值为 1。
	Number int
	// Size 是每页条数，对应 page_size；接口默认值为 20。
	Size int
}

// List 是分页响应；T 指定列表项类型，如 List[model.Food] 返回食物列表。
type List[T any] struct {
	// Items 是当前页的数据，空页通常返回空列表。
	Items []T `json:"items"`
	// Page 是本次返回的页码，从 1 开始。
	Page int `json:"page"`
	// PageSize 是请求的每页条数；最后一页的 Items 数量可以少于它。
	PageSize int `json:"page_size"`
	// Total 是符合筛选条件的总条数，不是当前页的条数或总页数。
	Total int64 `json:"total"`
}
