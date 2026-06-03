package shipment

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	createFn func(context.Context, CreateInput) (Request, error)
	updateFn func(context.Context, string, UpdateInput) (Request, error)
	listFn   func(context.Context) ([]Request, error)
	getFn    func(context.Context, string) (Request, error)
}

func (f fakeStore) EnsureSchema(context.Context) error { return nil }

func (f fakeStore) Create(ctx context.Context, input CreateInput) (Request, error) {
	return f.createFn(ctx, input)
}

func (f fakeStore) List(ctx context.Context) ([]Request, error) {
	return f.listFn(ctx)
}

func (f fakeStore) Get(ctx context.Context, id string) (Request, error) {
	return f.getFn(ctx, id)
}

func (f fakeStore) Update(ctx context.Context, id string, input UpdateInput) (Request, error) {
	return f.updateFn(ctx, id, input)
}

func TestServiceCreateRequiresCoreFields(t *testing.T) {
	svc := NewService(fakeStore{})

	_, err := svc.Create(context.Background(), CreateInput{})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestServiceCreateDefaultsStatusToNew(t *testing.T) {
	var got CreateInput
	svc := NewService(fakeStore{
		createFn: func(_ context.Context, input CreateInput) (Request, error) {
			got = input
			return Request{
				ID:             "req_01",
				Status:         StatusNew,
				CustomerName:   input.CustomerName,
				Contact:        input.Contact,
				CargoName:      input.CargoName,
				DestinationCity: input.DestinationCity,
				Comment:        input.Comment,
				CreatedAt:      time.Unix(1700000000, 0).UTC(),
				UpdatedAt:      time.Unix(1700000000, 0).UTC(),
			}, nil
		},
	})

	req, err := svc.Create(context.Background(), CreateInput{
		CustomerName:    "Ivan Petrov",
		Contact:         "+7 999 111-22-33",
		CargoName:       "Samples",
		DestinationCity: "Moscow",
		WeightKg:        floatPtr(12.5),
		Comment:         "Need delivery estimate",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if got.Status != StatusNew {
		t.Fatalf("expected store input status %q, got %q", StatusNew, got.Status)
	}
	if req.Status != StatusNew {
		t.Fatalf("expected created request status %q, got %q", StatusNew, req.Status)
	}
}

func TestServiceUpdateValidatesStatus(t *testing.T) {
	svc := NewService(fakeStore{})

	invalid := Status("unknown")
	_, err := svc.Update(context.Background(), "req_01", UpdateInput{Status: &invalid})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func floatPtr(v float64) *float64 { return &v }
