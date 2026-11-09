package service

import (
	"context"
	"strings"

	"dailymeal/backend/internal/agent"
	"dailymeal/backend/internal/config"
	"dailymeal/backend/internal/model"
	"dailymeal/backend/internal/nutrition"
)

// FoodImageInput 是通过一张截图新建或更新食物的请求参数。
type FoodImageInput struct {
	// ImageDataURL 是图片的 Base64 data URL，如 data:image/png;base64,...，不是文件路径。
	ImageDataURL string `json:"image_data_url"`
	// Name 是用户填写的名称；非空时覆盖识别名称，省略或仅有空白时使用识别结果。
	Name string `json:"name,omitempty"`
}

// FoodImageError 表示图片信息不足或与所选食物基准不一致，本次没有写入。
type FoodImageError string

// Error 返回可向调用方展示的识别信息错误。
func (e FoodImageError) Error() string { return string(e) }

// ImportFoodImage 识别截图，按数据库定义换算营养单位，再创建或更新食物。
func (s *Service) ImportFoodImage(ctx context.Context, id int64, in FoodImageInput) (*model.Food, error) {
	var existing *model.Food
	var err error
	if id != 0 {
		existing, err = s.store.Food(id)
		if err != nil {
			return nil, err
		}
	}
	cfg, err := config.LoadAgent()
	if err != nil {
		return nil, err
	}
	client := agent.Client{Config: cfg}
	reading, err := client.RecognizeFood(ctx, in.ImageDataURL)
	if err != nil {
		return nil, err
	}
	catalog, err := s.store.Nutrients()
	if err != nil {
		return nil, err
	}
	food, err := foodFromRecognition(reading, catalog)
	if err != nil {
		return nil, err
	}
	if name := strings.TrimSpace(in.Name); name != "" {
		food.Name = name
	}
	if existing == nil {
		if food.Name == "" || food.NutritionPer100["energy_kcal"] == nil {
			return nil, FoodImageError("新增食物需要名称和热量，请补填名称或使用包含热量的截图。")
		}
		// TODO：暂未对重复导入去重，重复提交同一截图会新增食物。
		return s.CreateFood(food)
	}
	if food.BasisUnit != existing.BasisUnit {
		return nil, FoodImageError("截图的重量或体积基准与所选食物不同，不能把 g 和 ml 互相转换。")
	}
	return s.PatchFood(id, FoodPatch{
		Name:            Field[string]{Set: food.Name != "", Value: food.Name},
		NutritionPer100: Field[nutrition.Values]{Set: true, Value: food.NutritionPer100},
	})
}

// foodFromRecognition 根据营养素目录将原始读数换算为每 100g/ml 的标准单位含量。
func foodFromRecognition(reading agent.FoodRecognition, catalog model.Nutrients) (FoodInput, error) {
	food := FoodInput{Name: strings.TrimSpace(reading.Name), NutritionPer100: nutrition.Values{}}
	if reading.BasisQuantity <= 0 {
		return food, FoodImageError("未识别到明确的营养份量，请使用标明每 100g、每 100ml 或具体份量的截图。")
	}
	switch reading.BasisUnit {
	case "g", "kg":
		food.BasisUnit = "g"
	case "ml", "L":
		food.BasisUnit = "ml"
	default:
		return food, FoodImageError("未识别到重量或体积单位，请使用带 g、kg、ml 或 L 的截图。")
	}
	factor := 100 / nutrition.Quantity(reading.BasisQuantity, reading.BasisUnit)
	definitions := make(map[string]model.Nutrient, len(catalog))
	for _, nutrient := range catalog {
		definitions[nutrient.Code] = nutrient
	}
	for _, item := range reading.Nutrients {
		nutrient, ok := definitions[item.Key]
		if !ok {
			return food, FoodImageError("识别结果包含不支持的营养指标，请重试或手动填写。")
		}
		value := item.Value
		if nutrient.Unit == "kcal" {
			switch item.Unit {
			case "kcal":
			case "kJ":
				value /= 4.184
			default:
				return food, FoodImageError("未识别到热量单位，请使用标明 kcal 或 kJ 的截图。")
			}
		} else {
			// 同一种营养素只做 g、mg、ug 的质量换算；不推断 IU、RAE 或 DFE 的口径。
			units := map[string]float64{"g": 1, "mg": 0.001, "ug": 0.000001}
			from, fromOK := units[item.Unit]
			to, toOK := units[nutrient.Unit]
			if !fromOK || !toOK {
				return food, FoodImageError("未识别到营养含量单位，请使用带 g、mg 或 μg 的截图。")
			}
			value *= from / to
		}
		food.NutritionPer100[nutrient.Key()] = nutrition.Number(value * factor)
	}
	if len(food.NutritionPer100) == 0 {
		return food, FoodImageError("没有识别到可保存的营养含量，请换一张清晰截图或手动填写。")
	}
	return food, nil
}
