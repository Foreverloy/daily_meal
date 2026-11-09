package service

import "encoding/json"

// Field 记录 PATCH 字段是否提交，保留 false、0 和营养值 null 的更新语义。
// 例如 {} 得到 Favorite.Set=false；{"favorite":false} 得到 Set=true、Value=false。
type Field[T any] struct {
	// Set 表示请求 JSON 中是否出现该字段，由 UnmarshalJSON 设置。
	Set bool
	// Value 是解析后的值；T 是具体类型，例如 Field[bool] 的 Value 为 bool。
	Value T
}

func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Set = true
	return json.Unmarshal(data, &f.Value)
}
