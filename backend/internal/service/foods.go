package service

import (
	"strings"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

type FoodInput struct {
	Name            string           `json:"name"`
	BasisUnit       string           `json:"basis_unit"`
	NutritionPer100 nutrition.Values `json:"nutrition_per_100"`
}

// FoodPatch 只更新提供的字段；favorite=false 取消收藏，省略时保留原状态。
type FoodPatch struct {
	Name            Field[string]           `json:"name"`
	NutritionPer100 Field[nutrition.Values] `json:"nutrition_per_100"`
	Favorite        Field[bool]             `json:"favorite"`
}

// CreateFood 新建未收藏的食物，收藏需要用户通过 PATCH 主动设置。
func (s *Service) CreateFood(in FoodInput) (*model.Food, error) {
	// TODO：暂未校验名称、营养基准和营养数值，调用方按接口约定提交。
	food := &model.Food{Name: strings.TrimSpace(in.Name), BasisUnit: in.BasisUnit, NutritionPer100: in.NutritionPer100.Normalized()}
	return food, s.store.CreateFood(food)
}

func (s *Service) Food(id int64) (*model.Food, error) {
	return s.store.Food(id)
}

func (s *Service) Foods(page Page) (List[model.Food], error) {
	result := List[model.Food]{Page: page.Number, PageSize: page.Size}
	var err error
	result.Items, result.Total, err = s.store.Foods(page.Query, page.Favorite, page.Number, page.Size)
	return result, err
}

// PatchFood 读取食物，只修改提交的字段，再保存。
func (s *Service) PatchFood(id int64, in FoodPatch) (*model.Food, error) {
	// TODO：暂未处理并发修改，同时更新同一食物时可能相互覆盖。
	food, err := s.store.Food(id)
	if err != nil {
		return nil, err
	}
	if in.Name.Set {
		food.Name = strings.TrimSpace(in.Name.Value)
	}
	if in.NutritionPer100.Set {
		for k, v := range in.NutritionPer100.Value {
			food.NutritionPer100[k] = v
		}
		food.NutritionPer100 = food.NutritionPer100.Normalized()
	}
	if in.Favorite.Set {
		food.Favorite = in.Favorite.Value
	}
	return food, s.store.SaveFood(food)
}
