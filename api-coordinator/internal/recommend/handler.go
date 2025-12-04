package recommend

import (
	"goflix/pkg/types"
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
	var req recommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido"})
		return
	}
	if req.TopN <= 0 {
		req.TopN = 10
	}
	results, err := h.svc.RecommendForUser(req.UserID, req.TopN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al calcular recomendaciones"})
		return
	}

	recommendations := make([]types.Neighbor, 0)
	for _, res := range results {
		recommendations = append(recommendations, res.Neighbors...)
	}

	// sort values
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Similarity > recommendations[j].Similarity
	})

	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}
