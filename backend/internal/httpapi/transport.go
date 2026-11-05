package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"dailymeal/backend/internal/agent"
	"dailymeal/backend/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func writeError(c *gin.Context, err error) {
	var dateError *time.ParseError
	if errors.As(err, &dateError) {
		badRequest(c, err)
		return
	}
	status, code, message := http.StatusInternalServerError, "internal_error", "服务器内部错误"
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status, code, message = http.StatusNotFound, "not_found", "资源不存在"
	} else if errors.Is(err, agent.ErrCall) {
		status, code, message = http.StatusBadGateway, "agent_error", "对话未完成，部分操作可能已经保存，请先查看相关饮食数据。"
		slog.Error("chat failed", "error", err)
	} else {
		slog.Error("request failed", "error", err)
	}
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message, "fields": gin.H{}}})
}

func badRequest(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": gin.H{
		"code": "bad_request", "message": err.Error(), "fields": gin.H{},
	}})
}

func respond(c *gin.Context, status int, value any, err error) {
	if err != nil {
		writeError(c, err)
		return
	}
	if status == http.StatusNoContent {
		c.Status(status)
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Data(status, "application/json; charset=utf-8", data)
}

func decode(c *gin.Context, target any) bool {
	// TODO：暂不校验未知字段、必填项和请求体大小，调用方须提供符合约定的输入。
	data, err := io.ReadAll(c.Request.Body)
	if err == nil {
		err = json.Unmarshal(data, target)
	}
	if err != nil {
		badRequest(c, err)
		return false
	}
	return true
}

func idParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequest(c, err)
		return 0, false
	}
	return id, true
}

func pageParams(c *gin.Context) (service.Page, bool) {
	p := service.Page{Query: c.Query("q"), Number: 1, Size: 20}
	var err error
	if raw, exists := c.GetQuery("page"); exists {
		p.Number, err = strconv.Atoi(raw)
		if err != nil {
			badRequest(c, err)
			return p, false
		}
	}
	if raw, exists := c.GetQuery("page_size"); exists {
		p.Size, err = strconv.Atoi(raw)
		if err != nil {
			badRequest(c, err)
			return p, false
		}
	}
	return p, true
}
