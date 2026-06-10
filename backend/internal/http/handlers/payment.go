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

// PaymentHandler — менеджерские ручки платежей по грузу. Клиентских ручек нет: платежи
// в MVP ведёт менеджер, клиент видит итог в переписке.
type PaymentHandler struct {
	payments *service.PaymentService
}

func NewPaymentHandler(payments *service.PaymentService) PaymentHandler {
	return PaymentHandler{payments: payments}
}

func (h PaymentHandler) Create(c *gin.Context) {
	managerID, ok := middleware.ManagerID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthorized", "authorization required")
		return
	}

	shipmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	var input service.CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	payment, err := h.payments.Create(c.Request.Context(), shipmentID, managerID, input)
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

	c.JSON(http.StatusCreated, payment)
}

func (h PaymentHandler) List(c *gin.Context) {
	shipmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}

	payments, err := h.payments.ListByShipment(c.Request.Context(), shipmentID)
	if err != nil {
		respondNotFoundOr500(c, err, "shipment not found")
		return
	}

	c.JSON(http.StatusOK, payments)
}

type updatePaymentStatusRequest struct {
	Status domain.PaymentStatus `json:"status"`
}

func (h PaymentHandler) UpdateStatus(c *gin.Context) {
	shipmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid shipment id")
		return
	}
	paymentID, err := uuid.Parse(c.Param("paymentID"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_id", "invalid payment id")
		return
	}

	var req updatePaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid_body", "invalid JSON body")
		return
	}

	payment, err := h.payments.UpdateStatus(c.Request.Context(), shipmentID, paymentID, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrValidation):
			respondError(c, http.StatusBadRequest, "validation", err.Error())
		case errors.Is(err, domain.ErrNotFound):
			respondError(c, http.StatusNotFound, "not_found", "payment not found")
		default:
			respondError(c, http.StatusInternalServerError, "internal", "internal server error")
		}
		return
	}

	c.JSON(http.StatusOK, payment)
}
