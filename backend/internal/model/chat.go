package model

// ToolResult 描述 Agent 返回的一次工具执行结果；定义这个结构不代表 Go 会执行工具。
type ToolResult struct {
	// Name 是被调用工具的名称，由 Agent 返回。
	Name string `json:"name"`
	// Arguments 是工具收到的参数对象；map[string]any 允许字符串、数字等不同值类型。
	Arguments map[string]any `json:"arguments"`
	// Result 是工具返回的 JSON 数据；any 允许对象、列表或简单值。
	Result any `json:"result"`
}

// ChatRequest 只保存单轮执行记录，暂不读取它来组织历史上下文。
// 模型凭据只从配置传给 Agent，不属于这条持久化记录。
type ChatRequest struct {
	// Base 嵌入本轮执行记录的主键和时间字段。
	Base
	// Message 是用户本轮提交的消息原文。
	Message string `json:"message"`
	// Reply 是 Agent 返回的回复文本。
	Reply string `json:"reply"`
	// ToolResults 是本轮返回的工具结果，以 JSONB 保存；纯文本对话时为空列表。
	ToolResults []ToolResult `json:"tool_results" gorm:"serializer:json;type:jsonb"`
	// Status 是执行状态：running（进行中）、completed（完成）、failed（失败）。
	Status string `json:"status"`
	// Error 保存失败原因，成功时为空；模型调用失败时使用 Go 生成的错误说明。
	Error string `json:"error"`
}
