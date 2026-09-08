// Package httpapi translates HTTP requests/responses; business decisions live in service.
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"dailymeal/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct{ service *service.Service }

func New(svc *service.Service, ping func(context.Context) error, openAPI []byte) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("request panic", "panic", recovered)
		writeError(c, errors.New("internal server error"))
	}))
	_ = r.SetTrustedProxies(nil)
	r.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/openapi.json", func(c *gin.Context) { c.Data(http.StatusOK, "application/json; charset=utf-8", openAPI) })
	r.NoRoute(func(c *gin.Context) { writeError(c, service.NotFound()) })
	h := &Handler{service: svc}
	a := r.Group("/api/v1")
	a.GET("/goal-settings/default", h.defaultGoal)
	a.PUT("/goal-settings/default", h.putDefaultGoal)
	a.GET("/daily-goals/:date", h.dailyGoal)
	a.PUT("/daily-goals/:date", h.putDailyGoal)
	a.DELETE("/daily-goals/:date", h.deleteDailyGoal)
	a.GET("/foods", h.foods)
	a.POST("/foods", h.createFood)
	a.GET("/foods/:id", h.food)
	a.PATCH("/foods/:id", h.patchFood)
	a.GET("/recipes", h.recipes)
	a.POST("/recipes", h.createRecipe)
	a.GET("/recipes/:id", h.recipe)
	a.PUT("/recipes/:id", h.putRecipe)
	a.GET("/meal-entries", h.entries)
	a.POST("/meal-entries", h.createEntry)
	a.GET("/meal-entries/:id", h.entry)
	a.PATCH("/meal-entries/:id", h.patchEntry)
	a.DELETE("/meal-entries/:id", h.deleteEntry)
	a.GET("/daily-summary", h.dailySummary)
	return r
}

func writeError(c *gin.Context, err error) {
	var appErr *service.Error
	status := http.StatusInternalServerError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case "validation_error":
			status = http.StatusUnprocessableEntity
		case "not_found":
			status = http.StatusNotFound
		case "bad_request":
			status = http.StatusBadRequest
		}
	} else {
		slog.ErrorContext(c.Request.Context(), "request failed", "error", err)
		appErr = &service.Error{Code: "internal_error", Message: "服务器内部错误", Fields: map[string]string{}}
	}
	c.AbortWithStatusJSON(status, gin.H{"error": appErr})
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
	c.JSON(status, value)
}

func decode(c *gin.Context, target any) bool {
	body := http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	data, err := io.ReadAll(body)
	if err == nil {
		data = bytes.TrimSpace(data)
		if len(data) == 0 || data[0] != '{' {
			err = errors.New("请求体必须是 JSON 对象")
		} else {
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.DisallowUnknownFields()
			err = decoder.Decode(target)
			if err == nil {
				var extra any
				if decoder.Decode(&extra) != io.EOF {
					err = errors.New("请求体只能包含一个 JSON 对象")
				}
			}
		}
	}
	if err != nil {
		writeError(c, &service.Error{Code: "bad_request", Message: "JSON 格式、字段或类型错误", Fields: map[string]string{"body": err.Error()}})
		return false
	}
	return true
}

func idParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(c, service.Invalid("id", "ID 须为正整数"))
		return 0, false
	}
	return id, true
}

func pageParams(c *gin.Context) (service.Page, bool) {
	p := service.Page{Query: c.Query("q"), Number: 1, Size: 20}
	for key, target := range map[string]*int{"page": &p.Number, "page_size": &p.Size} {
		if raw, exists := c.GetQuery(key); exists {
			value, err := strconv.Atoi(raw)
			if err != nil {
				writeError(c, service.Invalid(key, "须为整数"))
				return p, false
			}
			*target = value
		}
	}
	return p, true
}
