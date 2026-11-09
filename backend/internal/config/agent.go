package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AgentConfigPath 相对于 Go 服务的工作目录；本机 make run 在 backend 下运行。
const AgentConfigPath = "config/agent.yaml"

// ModelSettings 是 Python Agent 调用模型所需的配置；yaml 标签用于读取，json 标签用于转发。
type ModelSettings struct {
	// APIKey 是模型服务的访问凭据，仅随调用传给 Agent，不属于业务数据。
	APIKey string `json:"api_key" yaml:"api_key"`
	// BaseURL 是模型服务的 API 地址，如 http://localhost:9000/v1。
	BaseURL string `json:"base_url" yaml:"base_url"`
	// Model 是服务端识别的模型名称，具体值由所连接的模型服务决定。
	Model string `json:"model" yaml:"model"`
	// UseResponsesAPI 是传给 Agent 的接口选择开关，true 表示启用 Responses API。
	UseResponsesAPI bool `json:"use_responses_api" yaml:"use_responses_api"`
}

// AgentConfig 汇总 Agent 服务地址和模型配置，读取自 agent.yaml 的 default 节点。
type AgentConfig struct {
	// AgentURL 是 Python Agent 的 HTTP 地址，如 http://agent:8000。
	AgentURL string `yaml:"agent_url"`
	// ModelSettings 是嵌入字段；YAML 中 api_key、model 等与 agent_url 位于同一层。
	ModelSettings `yaml:",inline"`
}

// LoadAgent 读取 YAML 中的默认 Agent 配置，读取或解析失败时直接返回错误。
func LoadAgent() (AgentConfig, error) {
	// 用匿名结构体接收 YAML 的顶层对象。
	var document struct {
		// Default 对应 default: 节点，其他顶层节点不会被加载使用。
		Default AgentConfig `yaml:"default"`
	}
	data, err := os.ReadFile(AgentConfigPath)
	if err != nil {
		return document.Default, fmt.Errorf("读取 Agent 配置失败：%w", err)
	}
	if err := yaml.Unmarshal(data, &document); err != nil {
		// YAML 错误可能包含原始配置片段，不把凭据带入日志。
		return document.Default, fmt.Errorf("Agent 配置 YAML 格式错误")
	}
	return document.Default, nil
}
