// Package agent 调用 Python Agent；饮食业务仍由已有的 Go API 处理。
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
	"dailymeal/backend/internal/model"
)

var ErrCall = errors.New("Agent 调用失败")

type Client struct {
	Config config.AgentConfig
}

type Result struct {
	Reply       string             `json:"reply"`
	ToolResults []model.ToolResult `json:"tool_results"`
}

func (c Client) Chat(ctx context.Context, message string) (Result, error) {
	result := Result{ToolResults: []model.ToolResult{}}
	body, err := json.Marshal(struct {
		Message       string               `json:"message"`
		ModelSettings config.ModelSettings `json:"model_settings"`
	}{message, c.Config.ModelSettings})
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.Config.AgentURL, "/")+"/api/v1/chat", bytes.NewReader(body))
	if err != nil {
		return result, fmt.Errorf("%w：服务地址无效", ErrCall)
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := http.Client{Timeout: 3 * time.Minute}
	resp, err := httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("%w：连接中断或等待超时", ErrCall)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return result, fmt.Errorf("%w：响应无法解析", ErrCall)
		}
		return result, nil
	}
	if resp.StatusCode == http.StatusBadGateway {
		var failure struct {
			Error struct {
				ToolResults []model.ToolResult `json:"tool_results"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&failure); err != nil {
			return result, fmt.Errorf("%w：错误响应无法解析", ErrCall)
		}
		result.ToolResults = failure.Error.ToolResults
	}
	// 不保存或转发上游原始错误体，它可能带有请求中的模型凭据。
	return result, fmt.Errorf("%w：HTTP %d", ErrCall, resp.StatusCode)
}
