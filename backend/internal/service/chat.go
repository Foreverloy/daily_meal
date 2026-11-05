package service

import (
	"context"

	"dailymeal/backend/internal/agent"
	"dailymeal/backend/internal/config"
	"dailymeal/backend/internal/model"
)

type ChatInput struct {
	Message string `json:"message"`
}

func (s *Service) Chat(ctx context.Context, input ChatInput) (model.ChatRequest, error) {
	request := model.ChatRequest{
		Message: input.Message, Status: "running", ToolResults: []model.ToolResult{},
	}
	// TODO：暂不按请求去重，重复提交可能再次执行记餐等写操作。
	if err := s.store.CreateChat(&request); err != nil {
		return request, err
	}
	cfg, callErr := config.LoadAgent()
	if callErr == nil {
		client := agent.Client{Config: cfg}
		result, err := client.Chat(ctx, input.Message)
		request.Reply, request.ToolResults = result.Reply, result.ToolResults
		callErr = err
	}
	request.Status = "completed"
	if callErr != nil {
		request.Status, request.Error = "failed", callErr.Error()
	}
	// TODO：进程中断可能留下 running；连接中断时，已执行的工具结果可能没有返回。
	// 对话记录与工具的业务写入不共享事务，失败时先核对实际业务数据。
	if err := s.store.SaveChat(&request); err != nil {
		return request, err
	}
	return request, callErr
}
