// Package httpapi translates HTTP requests/responses; business decisions live in service.
package httpapi

import (
	"context"
	"net/http"

	"dailymeal/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler 负责解析 HTTP 请求和组织响应，把业务处理交给 Service。
type Handler struct {
	// service 是业务层入口，例如创建食物请求会调用 service.CreateFood。
	service *service.Service
}

// New 注册健康检查、接口文档及各业务路由。
func New(svc *service.Service, ping func(context.Context) error, openAPI []byte) *gin.Engine {
	r := gin.Default()
	_ = r.SetTrustedProxies(nil)
	r.GET("/healthz", func(c *gin.Context) {
		if err := ping(c.Request.Context()); err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/openapi.json", func(c *gin.Context) { c.Data(http.StatusOK, "application/json; charset=utf-8", openAPI) })
	r.NoRoute(func(c *gin.Context) { writeError(c, gorm.ErrRecordNotFound) })
	h := &Handler{service: svc}
	a := r.Group("/api/v1")
	a.GET("/nutrients", h.nutrients)
	a.POST("/chat", h.chat)
	a.GET("/goal-settings/default", h.defaultGoal)
	a.PUT("/goal-settings/default", h.putDefaultGoal)
	a.GET("/daily-goals/:date", h.dailyGoal)
	a.PUT("/daily-goals/:date", h.putDailyGoal)
	a.DELETE("/daily-goals/:date", h.deleteDailyGoal)
	a.GET("/foods", h.foods)
	a.POST("/foods", h.createFood)
	a.POST("/foods/from-image", h.createFoodFromImage)
	a.PATCH("/foods/:id/from-image", h.patchFoodFromImage)
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
