package recommend

import (
	"net/http"

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
	UserID string `json:"user_id" binding:"required"`
	TopN   int    `json:"top_n"`
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
	res, err := h.svc.RecommendForUser(req.UserID, req.TopN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al calcular recomendaciones"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recommendations": res})
}
