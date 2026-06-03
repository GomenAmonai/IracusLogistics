package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"iracus-logistic/backend/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(service *service.AuthService) AuthHandler {
	return AuthHandler{service: service}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	token, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondError(c, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
