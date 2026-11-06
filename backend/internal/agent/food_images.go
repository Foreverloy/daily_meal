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

type NutrientReading struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// FoodRecognition 保留图片中的原始基准和单位，供 Go 统一换算。
type FoodRecognition struct {
	Name          string            `json:"name"`
	BasisQuantity float64           `json:"basis_quantity"`
	BasisUnit     string            `json:"basis_unit"`
	Nutrients     []NutrientReading `json:"nutrients"`
}

func (c Client) RecognizeFood(ctx context.Context, imageDataURL string) (FoodRecognition, error) {
	var result FoodRecognition
	body, err := json.Marshal(struct {
		ImageDataURL  string               `json:"image_data_url"`
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
