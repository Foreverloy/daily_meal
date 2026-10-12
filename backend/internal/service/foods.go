package service

import (
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
	name, err := cleanName(in.Name)
	if err != nil {
		return nil, err
	}
	if in.BasisUnit != "g" && in.BasisUnit != "ml" {
		return nil, Invalid("basis_unit", "基准单位须为 g 或 ml")
	}
	if err := in.NutritionPer100.Validate(); err != nil {
		return nil, Invalid("nutrition_per_100", err.Error())
	}
	food := &model.Food{Name: name, BasisUnit: in.BasisUnit, NutritionPer100: in.NutritionPer100.Normalized()}
	return food, s.store.CreateFood(food)
}

func (s *Service) Food(id int64) (*model.Food, error) {
	food, err := s.store.Food(id, false)
	return food, resourceError(err)
}

func (s *Service) Foods(page Page) (List[model.Food], error) {
	result := List[model.Food]{Page: page.Number, PageSize: page.Size}
	if err := page.Validate(); err != nil {
		return result, err
	}
	err := s.transaction(func(tx *Service) error {
		var err error
		result.Items, result.Total, err = tx.store.Foods(page.Query, page.Favorite, page.Number, page.Size)
		return err
	})
	return result, err
}

// PatchFood 在同一事务中读取并锁定食物，校验通过后一次保存全部修改。
func (s *Service) PatchFood(id int64, in FoodPatch) (*model.Food, error) {
	var food *model.Food
	err := s.transaction(func(tx *Service) error {
		var err error
		food, err = tx.store.Food(id, true)
		if err != nil {
			return resourceError(err)
		}
		if in.Name.Set {
			if in.Name.Null {
				return Invalid("name", "名称不能为 null")
			}
			food.Name, err = cleanName(in.Name.Value)
			if err != nil {
				return err
			}
		}
		if in.NutritionPer100.Set {
			if in.NutritionPer100.Null {
				return Invalid("nutrition_per_100", "营养对象不能为 null")
			}
			for k, v := range in.NutritionPer100.Value {
				food.NutritionPer100[k] = v
			}
			if err := food.NutritionPer100.Validate(); err != nil {
				return Invalid("nutrition_per_100", err.Error())
			}
		}
		if in.Favorite.Set {
			if in.Favorite.Null {
				return Invalid("favorite", "收藏状态不能为 null")
			}
			food.Favorite = in.Favorite.Value
		}
		return tx.store.SaveFood(food)
	})
	return food, err
}
