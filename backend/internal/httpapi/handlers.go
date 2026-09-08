package httpapi

import (
	"net/http"

	"dailymeal/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) defaultGoal(c *gin.Context) {
	v, err := h.service.DefaultGoal(c.Request.Context())
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) putDefaultGoal(c *gin.Context) {
	var in service.GoalInput
	if !decode(c, &in) {
		return
	}
	v, err := h.service.PutDefaultGoal(c.Request.Context(), in)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) dailyGoal(c *gin.Context) {
	v, err := h.service.DailyGoal(c.Request.Context(), c.Param("date"))
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) putDailyGoal(c *gin.Context) {
	var in service.GoalInput
	if !decode(c, &in) {
		return
	}
	v, err := h.service.PutDailyGoal(c.Request.Context(), c.Param("date"), in)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) deleteDailyGoal(c *gin.Context) {
	err := h.service.DeleteDailyGoal(c.Request.Context(), c.Param("date"))
	respond(c, http.StatusNoContent, nil, err)
}

func (h *Handler) foods(c *gin.Context) {
	p, ok := pageParams(c)
	if !ok {
		return
	}
	v, err := h.service.Foods(c.Request.Context(), p)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) createFood(c *gin.Context) {
	var in service.FoodInput
	if !decode(c, &in) {
		return
	}
	v, err := h.service.CreateFood(c.Request.Context(), in)
	respond(c, http.StatusCreated, v, err)
}

func (h *Handler) food(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	v, err := h.service.Food(c.Request.Context(), id)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) patchFood(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in service.FoodPatch
	if !decode(c, &in) {
		return
	}
	v, err := h.service.PatchFood(c.Request.Context(), id, in)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) recipes(c *gin.Context) {
	p, ok := pageParams(c)
	if !ok {
		return
	}
	v, err := h.service.Recipes(c.Request.Context(), p)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) createRecipe(c *gin.Context) {
	var in service.RecipeInput
	if !decode(c, &in) {
		return
	}
	v, err := h.service.PutRecipe(c.Request.Context(), 0, in)
	respond(c, http.StatusCreated, v, err)
}

func (h *Handler) recipe(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	v, err := h.service.Recipe(c.Request.Context(), id)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) putRecipe(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in service.RecipeInput
	if !decode(c, &in) {
		return
	}
	v, err := h.service.PutRecipe(c.Request.Context(), id, in)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) entries(c *gin.Context) {
	v, err := h.service.Entries(c.Request.Context(), c.Query("date"), c.Query("meal_type"))
	respond(c, http.StatusOK, gin.H{"items": v}, err)
}

func (h *Handler) createEntry(c *gin.Context) {
	var in service.EntryInput
	if !decode(c, &in) {
		return
	}
	v, err := h.service.SaveEntry(c.Request.Context(), 0, in)
	respond(c, http.StatusCreated, v, err)
}

func (h *Handler) entry(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	v, err := h.service.Entry(c.Request.Context(), id)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) patchEntry(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in service.EntryInput
	if !decode(c, &in) {
		return
	}
	v, err := h.service.SaveEntry(c.Request.Context(), id, in)
	respond(c, http.StatusOK, v, err)
}

func (h *Handler) deleteEntry(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	err := h.service.DeleteEntry(c.Request.Context(), id)
	respond(c, http.StatusNoContent, nil, err)
}

func (h *Handler) dailySummary(c *gin.Context) {
	v, err := h.service.DailySummary(c.Request.Context(), c.Query("date"))
	respond(c, http.StatusOK, v, err)
}
