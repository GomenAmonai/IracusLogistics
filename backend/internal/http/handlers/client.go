package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"icaris-logistic/backend/internal/service"
)

type ClientHandler struct {
	service *service.ClientService
}

func NewClientHandler(service *service.ClientService) ClientHandler {
	return ClientHandler{service: service}
}

// List отдаёт менеджеру зарегистрированных клиентов — из этого списка он выбирает клиента
// при заведении груза.
func (h ClientHandler) List(c *gin.Context) {
	clients, err := h.service.List(c.Request.Context(), pageQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusOK, clients)
}
