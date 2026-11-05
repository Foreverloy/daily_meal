package model

type ToolResult struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
	Result    any            `json:"result"`
}

// ChatRequest 只保存单轮执行记录，暂不读取它来组织历史上下文。
// 模型凭据只从配置传给 Agent，不属于这条持久化记录。
type ChatRequest struct {
	Base
	Message     string       `json:"message"`
	Reply       string       `json:"reply"`
	ToolResults []ToolResult `json:"tool_results" gorm:"serializer:json;type:jsonb"`
	Status      string       `json:"status"`
	Error       string       `json:"error"`
}
