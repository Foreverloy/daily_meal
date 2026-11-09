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

// Client 通过 HTTP 调用 Python Agent，对话和截图识别共用这份连接配置。
type Client struct {
	// Config 包含 Agent 地址，以及每次请求转交给 Agent 的模型配置。
	Config config.AgentConfig
}

// Result 是 Agent 对话接口的返回内容，由 Service 写入对话执行记录。
type Result struct {
	// Reply 是展示给用户的回复文本。
	Reply string `json:"reply"`
	// ToolResults 是 Agent 返回的工具执行结果；纯文本对话时为空列表。
	ToolResults []model.ToolResult `json:"tool_results"`
}

// Chat 向 Python Agent 发送单轮消息，并返回回复及工具执行结果。
func (c Client) Chat(ctx context.Context, message string) (Result, error) {
	result := Result{ToolResults: []model.ToolResult{}}
	// 匿名结构体描述请求体，json 标签把 Go 字段名映射为接口字段名。
	body, err := json.Marshal(struct {
		// Message 是本轮用户输入的原文。
		Message string `json:"message"`
		// ModelSettings 是本轮交给 Agent 使用的模型配置，不写入对话记录。
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
		// 只读取错误响应中需要保留的工具结果，其余上游错误内容不转发。
		var failure struct {
			// Error 对应响应 JSON 最外层的 error 对象。
			Error struct {
				// ToolResults 是 Agent 在失败响应中仍然返回的工具结果。
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
