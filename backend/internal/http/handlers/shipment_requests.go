package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"iracus-logistic/backend/internal/shipment"
)

type ShipmentRequestHandler struct {
	service shipment.ServiceAPI
}

func NewShipmentRequestHandler(service shipment.ServiceAPI) ShipmentRequestHandler {
	return ShipmentRequestHandler{service: service}
}

func (h ShipmentRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input shipment.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, req)
}

func (h ShipmentRequestHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h ShipmentRequestHandler) Get(w http.ResponseWriter, r *http.Request) {
	req, err := h.service.Get(r.Context(), pathParam(r, "id"))
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, req)
}

func (h ShipmentRequestHandler) Update(w http.ResponseWriter, r *http.Request) {
	var input shipment.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	req, err := h.service.Update(r.Context(), pathParam(r, "id"), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, req)
}

func (h ShipmentRequestHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, shipment.ErrValidation):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, shipment.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func pathParam(r *http.Request, key string) string {
	if value := r.PathValue(key); value != "" {
		return value
	}

	return chi.URLParam(r, key)
}
