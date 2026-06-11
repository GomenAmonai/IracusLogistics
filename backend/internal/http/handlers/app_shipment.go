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

// AppShipmentHandler — клиентские ручки WebApp по грузам, чату и платежам. Доступ только
// к своим грузам: id клиента берётся из токена, не из запроса.
type AppShipmentHandler struct {
	shipments *service.ShipmentService
	messages  *service.MessageService
	payments  *service.PaymentService
}

func NewAppShipmentHandler(
	shipments *service.ShipmentService,
	messages *service.MessageService,
	payments *service.PaymentService,
) AppShipmentHandler {
	return AppShipmentHandler{shipments: shipments, messages: messages, payments: payments}
}

type sendMessageRequest struct {
	Text string `json:"text"`
}

func (h AppShipmentHandler) List(c *gin.Context) {
	clientID, ok := middleware.ClientID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	shipments, err := h.shipments.ListByClientID(c.Request.Context(), clientID, pageQuery(c))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		return
	}

	c.JSON(http.StatusOK, shipments)
}

func (h AppShipmentHandler) GetByID(c *gin.Context) {
	clientID, ok := middleware.ClientID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	detail, err := h.shipments.DetailForClient(c.Request.Context(), id, clientID)
	if err != nil {
		respondNotFoundOr500(c, err, "shipment not found")
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h AppShipmentHandler) ListMessages(c *gin.Context) {
	clientID, ok := middleware.ClientID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	messages, err := h.messages.ListForClient(c.Request.Context(), id, clientID)
	if err != nil {
		respondNotFoundOr500(c, err, "shipment not found")
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h AppShipmentHandler) ListPayments(c *gin.Context) {
	clientID, ok := middleware.ClientID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	payments, err := h.payments.ListForClient(c.Request.Context(), id, clientID)
	if err != nil {
		respondNotFoundOr500(c, err, "shipment not found")
		return
	}

	c.JSON(http.StatusOK, payments)
}

func (h AppShipmentHandler) SendMessage(c *gin.Context) {
	clientID, ok := middleware.ClientID(c)
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

	message, err := h.messages.SendFromClient(c.Request.Context(), id, clientID, req.Text)
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
