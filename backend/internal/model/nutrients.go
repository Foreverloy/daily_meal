package model

import "dailymeal/backend/internal/nutrition"

// Nutrient 定义一种营养指标及标准单位；同一条定义供所有食物共用。
type Nutrient struct {
	ID   int64  `json:"id" gorm:"primaryKey"`
	Code string `json:"code"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

// Key 返回接口与快照使用的指标名，如 protein 和 g 组成 protein_g。
func (n Nutrient) Key() string { return n.Code + "_" + n.Unit }

// Nutrients 是数据库中的营养素目录，计算时按这份目录统计全部指标。
type Nutrients []Nutrient

// Keys 返回目录中所有指标的接口名称，不在 Go 中另维护固定指标列表。
func (ns Nutrients) Keys() []string {
	keys := make([]string, 0, len(ns))
	for _, n := range ns {
		keys = append(keys, n.Key())
	}
	return keys
}

// FoodNutrient 保存食物每 100g/ml 的一项含量；nil 表示未知，指向 0 表示已知为零。
type FoodNutrient struct {
	FoodID       int64    `json:"food_id" gorm:"primaryKey;autoIncrement:false"`
	NutrientID   int64    `json:"nutrient_id" gorm:"primaryKey;autoIncrement:false"`
	AmountPer100 *float64 `json:"amount_per_100" gorm:"column:amount_per_100"`
}

// SetNutrition 将接口含量转换成关联记录，并按目录补齐未知指标，供保存与返回使用。
func (f *Food) SetNutrition(catalog Nutrients, values nutrition.Values) {
	f.NutritionPer100 = values.Normalized(catalog.Keys())
	f.Nutrients = make([]FoodNutrient, 0, len(catalog))
	for _, n := range catalog {
		f.Nutrients = append(f.Nutrients, FoodNutrient{
			FoodID: f.ID, NutrientID: n.ID, AmountPer100: f.NutritionPer100[n.Key()],
		})
	}
}

// LoadNutrition 将已加载的关联记录组织成接口含量；新指标尚无关联时按未知返回。
func (f *Food) LoadNutrition(catalog Nutrients) {
	amounts := make(map[int64]*float64, len(f.Nutrients))
	for _, row := range f.Nutrients {
		amounts[row.NutrientID] = row.AmountPer100
	}
	f.NutritionPer100 = make(nutrition.Values, len(catalog))
	for _, n := range catalog {
		f.NutritionPer100[n.Key()] = amounts[n.ID]
	}
}
