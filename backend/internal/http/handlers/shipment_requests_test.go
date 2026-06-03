package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"iracus-logistic/backend/internal/shipment"
)

type stubShipmentService struct {
	createFn func(shipment.CreateInput) (shipment.Request, error)
	listFn   func() ([]shipment.Request, error)
	getFn    func(string) (shipment.Request, error)
	updateFn func(string, shipment.UpdateInput) (shipment.Request, error)
}

func (s stubShipmentService) Create(_ context.Context, input shipment.CreateInput) (shipment.Request, error) {
	return s.createFn(input)
}

func (s stubShipmentService) List(_ context.Context) ([]shipment.Request, error) {
	return s.listFn()
}

func (s stubShipmentService) Get(_ context.Context, id string) (shipment.Request, error) {
	return s.getFn(id)
}

func (s stubShipmentService) Update(_ context.Context, id string, input shipment.UpdateInput) (shipment.Request, error) {
	return s.updateFn(id, input)
}

func TestShipmentHandlerCreate(t *testing.T) {
	handler := NewShipmentRequestHandler(stubShipmentService{
		createFn: func(input shipment.CreateInput) (shipment.Request, error) {
			return shipment.Request{
				ID:              "req_01",
				Status:          shipment.StatusNew,
				CustomerName:    input.CustomerName,
				Contact:         input.Contact,
				CargoName:       input.CargoName,
				DestinationCity: input.DestinationCity,
				Comment:         input.Comment,
				CreatedAt:       time.Unix(1700000000, 0).UTC(),
				UpdatedAt:       time.Unix(1700000000, 0).UTC(),
			}, nil
		},
	})

	body := bytes.NewBufferString(`{"customer_name":"Ivan","contact":"+79991112233","cargo_name":"Samples","destination_city":"Moscow","comment":"Estimate needed","weight_kg":12.5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/shipment-requests", body)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var created shipment.Request
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if created.ID != "req_01" {
		t.Fatalf("expected request id req_01, got %q", created.ID)
	}
}

func TestShipmentHandlerList(t *testing.T) {
	handler := NewShipmentRequestHandler(stubShipmentService{
		listFn: func() ([]shipment.Request, error) {
			return []shipment.Request{{ID: "req_01", Status: shipment.StatusNew}}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/shipment-requests", nil)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var listed []shipment.Request
	if err := json.NewDecoder(rr.Body).Decode(&listed); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 shipment request, got %d", len(listed))
	}
}

func TestShipmentHandlerGet(t *testing.T) {
	handler := NewShipmentRequestHandler(stubShipmentService{
		getFn: func(id string) (shipment.Request, error) {
			return shipment.Request{ID: id, Status: shipment.StatusNew}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/shipment-requests/req_01", nil)
	req.SetPathValue("id", "req_01")
	rr := httptest.NewRecorder()

	handler.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var fetched shipment.Request
	if err := json.NewDecoder(rr.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if fetched.ID != "req_01" {
		t.Fatalf("expected request id req_01, got %q", fetched.ID)
	}
}

func TestShipmentHandlerUpdate(t *testing.T) {
	handler := NewShipmentRequestHandler(stubShipmentService{
		updateFn: func(id string, input shipment.UpdateInput) (shipment.Request, error) {
			status := shipment.StatusNew
			if input.Status != nil {
				status = *input.Status
			}

			return shipment.Request{
				ID:             id,
				Status:         status,
				ManagerComment: input.ManagerComment,
				CreatedAt:      time.Unix(1700000000, 0).UTC(),
				UpdatedAt:      time.Unix(1700000060, 0).UTC(),
			}, nil
		},
	})

	body := bytes.NewBufferString(`{"status":"in_calculation","manager_comment":"Need supplier quote"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/shipment-requests/req_01", body)
	req.SetPathValue("id", "req_01")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var updated shipment.Request
	if err := json.NewDecoder(rr.Body).Decode(&updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if updated.Status != shipment.StatusInCalculation {
		t.Fatalf("expected status %q, got %q", shipment.StatusInCalculation, updated.Status)
	}
}
