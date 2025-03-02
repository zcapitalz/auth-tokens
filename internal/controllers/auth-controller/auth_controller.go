package authcontroller

import (
	"auth/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	authService AuthService
}

type AuthService interface {
	CreateSession(userID uuid.UUID, requestIP string) (*domain.Session, error)
	RefreshSession(session *domain.Session, requestIP string) (*domain.Session, error)
}

func NewAuthController(authService AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) RegisterRoutes(engine *gin.Engine) {
	sessionGroup := engine.Group("sessions")
	sessionGroup.POST("", c.createSession)
	sessionGroup.POST("/refresh", c.refreshSession)
}
