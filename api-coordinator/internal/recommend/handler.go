package recommend

import (
	"goflix/pkg/types"
	"log"
	"net/http"
	"sort"

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
	log.Println("Recomendaciones calculadas")
	log.Println("Recomendaciones: ", recommendations)
	// sort values
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Similarity > recommendations[j].Similarity
	})

	c.JSON(http.StatusOK, gin.H{"recommendations": recommendations})
}
