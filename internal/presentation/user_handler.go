// presentation/user_handler.go
package presentation

import (
	"go-ddd/internal/application"
	"go-ddd/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	service *application.UserService
}

func NewUserHandler(service *application.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	var user domain.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.service.RegisterUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *UserHandler) AuthenticateUser(c *gin.Context) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.service.AuthenticateUser(credentials.Username, credentials.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
