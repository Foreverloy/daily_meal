package model

import (
	"database/sql/driver"
	"fmt"
	"time"

	"dailymeal/backend/internal/nutrition"
)

// Date 表示不含时分秒的业务日期，如 "2026-11-09"，对应 PostgreSQL 的 date 类型。
type Date string

func (d *Date) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		*d = Date(v.Format(time.DateOnly))
	case string:
		*d = Date(v)
	case []byte:
		*d = Date(string(v))
	default:
		return fmt.Errorf("invalid database date: %T", value)
	}
	return nil
}
func (d Date) Value() (driver.Value, error) { return string(d), nil }

// Base 是各数据库模型共用的字段，通过嵌入复用；外层对象可直接访问 ID 等字段。
// JSON 中 id、created_at、updated_at 直接出现在外层，不会包在 base 对象中。
type Base struct {
	// ID 是数据库主键，用于接口路径和模型关联，如 /foods/12 中的 12。
	ID int64 `json:"id" gorm:"primaryKey"`
	// CreatedAt 是记录创建时间，由 GORM 自动填充。
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 是记录最近更新时间，由 GORM 在保存时自动维护。
	UpdatedAt time.Time `json:"updated_at"`
}

// GoalValues 保存每日热量目标和三大营养素的供能占比，占比用 0～100 的百分数表示。
// 例如 2000 kcal、蛋白质占 20%，换算后的蛋白质目标为 2000 × 20% ÷ 4 = 100g。
type GoalValues struct {
	// EnergyKcal 是每日总热量目标，单位为 kcal（大卡）。
	EnergyKcal float64 `json:"energy_kcal"`
	// ProteinPercent 是蛋白质供能占比，如 20 表示 20%，不是 20g。
	ProteinPercent float64 `json:"protein_percent"`
	// CarbohydratePercent 是碳水化合物供能占比，如 50 表示 50%。
	CarbohydratePercent float64 `json:"carbohydrate_percent"`
	// FatPercent 是脂肪供能占比；三个占比按业务约定合计为 100。
	FatPercent float64 `json:"fat_percent"`
}

func (g GoalValues) Targets() map[string]float64 {
	return map[string]float64{
		"energy_kcal":    g.EnergyKcal,
		"protein_g":      g.EnergyKcal * (g.ProteinPercent / 100) / 4,
		"carbohydrate_g": g.EnergyKcal * (g.CarbohydratePercent / 100) / 4,
		"fat_g":          g.EnergyKcal * (g.FatPercent / 100) / 9,
	}
}

// GoalSetting 是一版默认目标，从生效日起沿用，直到有后续默认目标接替。
type GoalSetting struct {
	// Base 嵌入这条目标记录的主键和时间字段。
	Base
	// EffectiveOn 是生效日期，查询某天时选取不晚于该日的最新一版。
	EffectiveOn Date `json:"effective_on" gorm:"type:date"`
	// GoalValues 嵌入热量和供能占比，作为同一张表的字段保存。
	GoalValues `gorm:"embedded"`
}

// DailyGoal 保存某一天单独设置的目标，仅在该日覆盖默认目标。
type DailyGoal struct {
	// Base 嵌入这条单日目标的主键和时间字段。
	Base
	// Date 是覆盖目标适用的日期，如 "2026-11-09"。
	Date Date `json:"date" gorm:"type:date"`
	// GoalValues 嵌入当天使用的完整热量和供能占比。
	GoalValues `gorm:"embedded"`
}

// Food 是食物库中的一项食物，营养值统一按每 100g 或每 100ml 保存。
type Food struct {
	// Base 嵌入食物的主键和时间字段。
	Base
	// Name 是展示和搜索使用的名称，如“纯牛奶”。
	Name string `json:"name"`
	// BasisUnit 是营养基准单位，只使用 g 或 ml；ml 表示下面的值均按每 100ml 计。
	BasisUnit string `json:"basis_unit"`
	// NutritionPer100 是每 100 个基准单位的营养，如 {"energy_kcal":60,"protein_g":3.2}。
	NutritionPer100 nutrition.Values `json:"nutrition_per_100" gorm:"column:nutrition_per_100;serializer:json;type:jsonb"`
	// Favorite 是食物库的收藏状态，不参与营养计算或饮食记录快照。
	Favorite bool `json:"favorite"`
}

// Recipe 保存菜品模板的基本信息，食材组成单独放在 RecipeIngredient 中。
type Recipe struct {
	// Base 嵌入菜品的主键和时间字段。
	Base
	// Name 是菜品名称，如“鸡胸肉沙拉”。
	Name string `json:"name"`
}

// RecipeIngredient 表示某道菜使用的一项食物及用量。
type RecipeIngredient struct {
	// Base 嵌入食材关联记录自身的主键和时间字段。
	Base
	// RecipeID 指向所属菜品模板的 ID。
	RecipeID int64 `json:"recipe_id"`
	// FoodID 指向所用食物的 ID。
	FoodID int64 `json:"food_id"`
	// Quantity 是已换算成 g/ml 的用量，单位跟随食物；输入 0.2kg 时保存 200。
	Quantity float64 `json:"quantity"`
	// Food 是通过 FoodID 预加载的食物对象，供营养计算使用；json:"-" 表示不直接输出。
	Food Food `json:"-" gorm:"foreignKey:FoodID"`
}

// IngredientSnapshot 保存计算时读取的一项食材数据，用于菜品展示和饮食记录快照。
type IngredientSnapshot struct {
	// FoodID 是来源食物的 ID，保留关联以便识别食材。
	FoodID int64 `json:"food_id"`
	// Name 是生成快照时的食物名称。
	Name string `json:"name"`
	// BasisUnit 是生成快照时的食物基准单位：g 或 ml。
	BasisUnit string `json:"basis_unit"`
	// Quantity 是整道菜中该食材的用量，使用 BasisUnit，如 150 表示 150g 或 150ml。
	Quantity float64 `json:"quantity"`
	// NutritionPer100 是当时每 100g/100ml 的营养，写入饮食记录后不随食物库更新。
	NutritionPer100 nutrition.Values `json:"nutrition_per_100"`
}

// Snapshot 保存记餐时的名称和营养依据，历史记录的展示与计算使用这份数据。
// 食物记录填写 BasisUnit、NutritionPer100；菜品记录填写 Ingredients。
type Snapshot struct {
	// Name 是当时的食物或菜品名称。
	Name string `json:"name"`
	// BasisUnit 是食物记录的 g/ml 基准；菜品记录为空，omitempty 使其不出现在 JSON 中。
	BasisUnit string `json:"basis_unit,omitempty"`
	// NutritionPer100 是食物记录的每 100g/100ml 营养，用于修改用量后重新计算。
	NutritionPer100 nutrition.Values `json:"nutrition_per_100,omitempty"`
	// Ingredients 是菜品记录的完整食材快照，食物记录不使用此字段。
	Ingredients []IngredientSnapshot `json:"ingredients,omitempty"`
}

// MealEntry 是某天某餐的一条饮食记录，保存当时快照和按实际用量计算的营养。
type MealEntry struct {
	// Base 嵌入饮食记录的主键和时间字段。
	Base
	// Date 是实际用餐日期，可与记录创建日期不同，用于补记饮食。
	Date Date `json:"date" gorm:"type:date"`
	// MealType 是餐次：breakfast（早）、lunch（中）、dinner（晚）。
	MealType string `json:"meal_type"`
	// ItemType 表示记录对象：food 为食物，recipe 为整道菜。
	ItemType string `json:"item_type"`
	// FoodID 是食物记录的来源 ID；菜品记录为 nil，输出 JSON null。
	FoodID *int64 `json:"food_id"`
	// RecipeID 是菜品记录的来源 ID；食物记录为 nil。
	RecipeID *int64 `json:"recipe_id"`
	// Quantity 是食物的 g/ml 用量，单位见 Snapshot.BasisUnit；菜品整道计入，此处为 nil。
	Quantity *float64 `json:"quantity"`
	// Snapshot 是记餐时保存的数据副本，食物库或菜品模板后续变更不会自动改写它。
	Snapshot Snapshot `json:"snapshot" gorm:"serializer:json;type:jsonb"`
	// NutritionTotals 是本条记录的营养合计，供餐次和全天汇总直接使用。
	NutritionTotals nutrition.Totals `json:"nutrition_totals" gorm:"serializer:json;type:jsonb"`
}
