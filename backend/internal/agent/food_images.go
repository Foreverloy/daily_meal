package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"dailymeal/backend/internal/config"
)

var ErrRecognition = errors.New("截图识别失败")

// NutrientReading 表示截图中的一项原始营养读数，尚未换算成食物库单位。
// 例如：{"key":"energy","value":420,"unit":"kJ"} 表示热量为 420 千焦。
type NutrientReading struct {
	// Key 是营养指标名，如 energy、protein、sodium；这里不带单位后缀。
	Key string `json:"key"`
	// Value 是图片标注的数值；0 表示明确为零，未识别的指标不放入列表。
	Value float64 `json:"value"`
	// Unit 是该读数的单位，如 kcal、kJ、g、mg、ug（微克）。
	Unit string `json:"unit"`
}

// FoodRecognition 保留图片中的原始基准和单位，供 Go 统一换算。
// 例如截图写“每 250ml 牛奶含蛋白质 8g”，则 BasisQuantity=250、BasisUnit="ml"，
// Nutrients 中包含 {Key: "protein", Value: 8, Unit: "g"}；保存时再换成每 100ml 含 3.2g。
type FoodRecognition struct {
	// Name 是识别出的食物名称，如“纯牛奶”；图片没有名称时可为空。
	Name string `json:"name"`
	// BasisQuantity 是图片营养表对应的食物份量，如“每 250ml”中的 250。
	BasisQuantity float64 `json:"basis_quantity"`
	// BasisUnit 是上述份量的单位：g、kg、ml 或 L，不是营养素本身的单位。
	BasisUnit string `json:"basis_unit"`
	// Nutrients 是这份食物的营养读数列表，未识别到的指标不包含在内。
	Nutrients []NutrientReading `json:"nutrients"`
}

// RecognizeFood 将图片和模型配置发给 Python Agent，并读取原始识别结果。
func (c Client) RecognizeFood(ctx context.Context, imageDataURL string) (FoodRecognition, error) {
	var result FoodRecognition
	// 匿名结构体只用于这次请求；json 标签指定发给 Agent 的 JSON 字段名。
	body, err := json.Marshal(struct {
		// ImageDataURL 是带图片类型前缀的 Base64 内容，如 data:image/png;base64,...。
		ImageDataURL string `json:"image_data_url"`
		// ModelSettings 是 Agent 调用模型时使用的配置，包含服务地址和凭据。
		ModelSettings config.ModelSettings `json:"model_settings"`
	}{imageDataURL, c.Config.ModelSettings})
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.Config.AgentURL, "/")+"/api/v1/food-recognitions", bytes.NewReader(body))
	if err != nil {
		return result, fmt.Errorf("%w：服务地址无效", ErrRecognition)
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{Timeout: 3 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("%w：连接中断或等待超时", ErrRecognition)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// 不转发原始错误体，它可能带有模型凭据。
		return result, fmt.Errorf("%w：HTTP %d", ErrRecognition, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("%w：响应无法解析", ErrRecognition)
	}
	return result, nil
}
