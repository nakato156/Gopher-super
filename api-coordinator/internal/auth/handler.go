package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// authRequest representa el body de login/register.
type authRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// authResponse es la respuesta estándar.
type authResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// Handler expone los endpoints HTTP de auth.
type Handler struct {
	svc     Service
	timeout time.Duration
}

// NewHandler crea un handler de autenticación.
func NewHandler(svc Service) *Handler {
	return &Handler{
		svc:     svc,
		timeout: 5 * time.Second,
	}
}

// Register registra las rutas /api/auth/*
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.login)
	rg.POST("/register", h.register)
}

func (h *Handler) register(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	userID, token, err := h.svc.Register(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "el usuario ya existe"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo registrar", "msg": err})
		}
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		UserID: userID,
		Token:  token,
	})
}

func (h *Handler) login(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload inválido", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	userID, token, err := h.svc.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales inválidas"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo iniciar sesión"})
		}
		return
	}

	c.JSON(http.StatusOK, authResponse{
		UserID: userID,
		Token:  token,
	})
}
