package service

import "encoding/json"

// Field 记录 PATCH 字段是否提交，保留 false、0 和营养值 null 的更新语义。
type Field[T any] struct {
	Set   bool
	Value T
}

func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Set = true
	return json.Unmarshal(data, &f.Value)
}
