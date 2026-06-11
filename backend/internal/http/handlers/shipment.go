package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"icaris-logistic/backend/internal/domain"
	"icaris-logistic/backend/internal/middleware"
	"icaris-logistic/backend/internal/service"
)

// ShipmentHandler — менеджерские ручки по грузам и переписке. Менеджер видит все грузы.
type ShipmentHandler struct {
	shipments *service.ShipmentService
	messages  *service.MessageService
}

func NewShipmentHandler(shipments *service.ShipmentService, messages *service.MessageService) ShipmentHandler {
	return ShipmentHandler{shipments: shipments, messages: messages}
}

func (h ShipmentHandler) Create(c *gin.Context) {
	managerID, ok := middleware.ManagerID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	var input service.CreateShipmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	shipment, err := h.shipments.Create(c.Request.Context(), managerID, input)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			respondError(c, http.StatusBadRequest, "validation", err.Error())
			return
		}
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusCreated, shipment)
}

func (h ShipmentHandler) List(c *gin.Context) {
	shipments, err := h.shipments.List(c.Request.Context(), pageQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusOK, shipments)
}

func (h ShipmentHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	detail, err := h.shipments.Detail(c.Request.Context(), id)
	if err != nil {
		respondNotFoundOr500(c, err, "shipment not found")
		return
	}

	c.JSON(http.StatusOK, detail)
}

type updateShipmentStatusRequest struct {
	Status  domain.ShipmentStatus `json:"status"`
	Comment string                `json:"comment"`
}

func (h ShipmentHandler) UpdateStatus(c *gin.Context) {
	managerID, ok := middleware.ManagerID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	var req updateShipmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	shipment, err := h.shipments.UpdateStatus(c.Request.Context(), id, managerID, req.Status, req.Comment)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			respondError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, domain.ErrNotFound):
			respondError(c, http.StatusNotFound, "not_found", "shipment not found")
		default:
			respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, shipment)
}

func (h ShipmentHandler) ListMessages(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	messages, err := h.messages.ListForManager(c.Request.Context(), id)
	if err != nil {
		respondNotFoundOr500(c, err, "shipment not found")
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h ShipmentHandler) SendMessage(c *gin.Context) {
	managerID, ok := middleware.ManagerID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	message, err := h.messages.SendFromManager(c.Request.Context(), id, managerID, req.Text)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			respondError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, domain.ErrNotFound):
			respondError(c, http.StatusNotFound, "not_found", "shipment not found")
		default:
			respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		}
		return
	}

	c.JSON(http.StatusCreated, message)
}
