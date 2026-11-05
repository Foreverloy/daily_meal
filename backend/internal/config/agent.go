package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AgentConfigPath 相对于 Go 服务的工作目录；本机 make run 在 backend 下运行。
const AgentConfigPath = "config/agent.yaml"

type ModelSettings struct {
	APIKey          string `json:"api_key" yaml:"api_key"`
	BaseURL         string `json:"base_url" yaml:"base_url"`
	Model           string `json:"model" yaml:"model"`
	UseResponsesAPI bool   `json:"use_responses_api" yaml:"use_responses_api"`
}

type AgentConfig struct {
	AgentURL      string `yaml:"agent_url"`
	ModelSettings `yaml:",inline"`
}

func LoadAgent() (AgentConfig, error) {
	var document struct {
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
