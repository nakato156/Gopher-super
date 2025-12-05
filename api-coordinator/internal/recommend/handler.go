package recommend

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers recommend endpoints on the provided router group.
// It registers POST "" so that mounting on a group like r.Group("/recomend")
// exposes POST /recomend
func (h *Handler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("", h.Recommend)
	g.GET("/popular", h.GetPopularMovies)
}

func (h *Handler) GetPopularMovies(c *gin.Context) {
	topN := 10
	if nStr := c.Query("top_n"); nStr != "" {
		if n, err := strconv.Atoi(nStr); err == nil && n > 0 {
			topN = n
		}
	}

	movies, err := h.svc.GetPopularMovies(c.Request.Context(), topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching popular movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

type recommendRequest struct {
	UserID int `json:"user_id" binding:"required"`
	TopN   int `json:"top_n"`
}

func (h *Handler) Recommend(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id format"})
		return
	}

	var req recommendRequest
	// Optional: still bind for TopN if provided, but ignore UserID in body
	if err := c.ShouldBindJSON(&req); err == nil {
		// If body is present, use TopN from it
	}

	if req.TopN <= 0 {
		req.TopN = 10
	}

	recommendations, err := h.svc.GetRecommendationsWithDetails(c.Request.Context(), userID, req.TopN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al calcular recomendaciones"})
		return
	}

	// sort values
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})

	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}
