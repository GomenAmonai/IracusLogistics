package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/service"
)

type LeadHandler struct {
	service *service.LeadService
}

func NewLeadHandler(service *service.LeadService) LeadHandler {
	return LeadHandler{service: service}
}

func (h LeadHandler) Create(c *gin.Context) {
	var input service.CreateLeadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	lead, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			respondError(c, http.StatusBadRequest, "validation", err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusCreated, lead)
}

func (h LeadHandler) List(c *gin.Context) {
	leads, err := h.service.List(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusOK, leads)
}

func (h LeadHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid lead id")
		return
	}

	lead, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondError(c, http.StatusNotFound, "not_found", "lead not found")
			return
		}
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusOK, lead)
}

type updateLeadStatusRequest struct {
	Status domain.LeadStatus `json:"status"`
}

func (h LeadHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid lead id")
		return
	}

	var req updateLeadStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	lead, err := h.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			respondError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, domain.ErrNotFound):
			respondError(c, http.StatusNotFound, "not_found", "lead not found")
		default:
			respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, lead)
}
