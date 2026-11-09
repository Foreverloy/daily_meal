package service

import (
	"errors"
	"strings"

	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

// FoodInput 是创建食物的参数；截图识别结果也会先转换成这个结构再保存。
type FoodInput struct {
	// Name 是食物名称，如“纯牛奶”。
	Name string `json:"name"`
	// BasisUnit 是营养基准单位：g 或 ml；如 ml 表示营养按每 100ml 输入。
	BasisUnit string `json:"basis_unit"`
	// NutritionPer100 是每 100g/100ml 的营养；未提供的指标在保存时补为未知值 nil。
	NutritionPer100 nutrition.Values `json:"nutrition_per_100"`
}

// FoodPatch 只更新提供的字段；favorite=false 取消收藏，省略时保留原状态。
type FoodPatch struct {
	// Name 是要替换的名称，省略时保留原名称。
	Name Field[string] `json:"name"`
	// NutritionPer100 按指标逐项合并，如 {"protein_g":0} 设为零，{"sodium_mg":null} 清为未知。
	NutritionPer100 Field[nutrition.Values] `json:"nutrition_per_100"`
	// Favorite 是收藏开关；Field 使省略该字段与明确传 false 能够区分。
	Favorite Field[bool] `json:"favorite"`
}

// CreateFood 新建未收藏的食物，收藏需要用户通过 PATCH 主动设置。
func (s *Service) CreateFood(in FoodInput) (*model.Food, error) {
	// TODO：暂未校验名称、营养基准和营养数值，调用方按接口约定提交。
	// 原 foods JSONB 列要求热量非空，拆表后在此保留同一要求。
	if in.NutritionPer100["energy_kcal"] == nil {
		return nil, errors.New("food energy is required")
	}
	catalog, err := s.store.Nutrients()
	if err != nil {
		return nil, err
	}
	food := &model.Food{Name: strings.TrimSpace(in.Name), BasisUnit: in.BasisUnit}
	food.SetNutrition(catalog, in.NutritionPer100)
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
		if food.NutritionPer100["energy_kcal"] == nil {
			return nil, errors.New("food energy is required")
		}
		catalog, err := s.store.Nutrients()
		if err != nil {
			return nil, err
		}
		food.SetNutrition(catalog, food.NutritionPer100)
	}
	if in.Favorite.Set {
		food.Favorite = in.Favorite.Value
	}
	return food, s.store.SaveFood(food)
}
