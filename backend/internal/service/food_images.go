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

func (e FoodImageError) Error() string { return string(e) }

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
	food, err := foodFromRecognition(reading)
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

func foodFromRecognition(reading agent.FoodRecognition) (FoodInput, error) {
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
	for _, item := range reading.Nutrients {
		var key string
		var value float64
		if item.Key == "energy" {
			key, value = "energy_kcal", item.Value
			switch item.Unit {
			case "kcal":
			case "kJ":
				value /= 4.184
			default:
				return food, FoodImageError("未识别到热量单位，请使用标明 kcal 或 kJ 的截图。")
			}
		} else {
			// 所有质量单位先换成 g，再换到食物库约定的营养素单位。
			units := map[string]float64{"g": 1, "mg": 0.001, "ug": 0.000001}
			unit, ok := units[item.Unit]
			if !ok {
				return food, FoodImageError("未识别到营养含量单位，请使用带 g、mg 或 μg 的截图。")
			}
			value = item.Value * unit
			switch item.Key {
			case "protein", "carbohydrate", "fat":
				key = item.Key + "_g"
			case "sodium", "calcium", "vitamin_c":
				key, value = item.Key+"_mg", value*1000
			case "vitamin_d":
				key, value = "vitamin_d_ug", value*1000000
			default:
				return food, FoodImageError("识别结果包含不支持的营养指标，请重试或手动填写。")
			}
		}
		food.NutritionPer100[key] = nutrition.Number(value * factor)
	}
	if len(food.NutritionPer100) == 0 {
		return food, FoodImageError("没有识别到可保存的营养含量，请换一张清晰截图或手动填写。")
	}
	return food, nil
}
