package shipment

import (
	"context"
	"fmt"
	"strings"
)

type Store interface {
	EnsureSchema(ctx context.Context) error
	Create(ctx context.Context, input CreateInput) (Request, error)
	List(ctx context.Context) ([]Request, error)
	Get(ctx context.Context, id string) (Request, error)
	Update(ctx context.Context, id string, input UpdateInput) (Request, error)
}

type ServiceAPI interface {
	Create(ctx context.Context, input CreateInput) (Request, error)
	List(ctx context.Context) ([]Request, error)
	Get(ctx context.Context, id string) (Request, error)
	Update(ctx context.Context, id string, input UpdateInput) (Request, error)
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Request, error) {
	input = input.normalized()
	input.Status = StatusNew

	if err := validateCreateInput(input); err != nil {
		return Request{}, err
	}

	return s.store.Create(ctx, input)
}

func (s *Service) List(ctx context.Context) ([]Request, error) {
	return s.store.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (Request, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Request{}, fmt.Errorf("%w: id is required", ErrValidation)
	}

	return s.store.Get(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (Request, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Request{}, fmt.Errorf("%w: id is required", ErrValidation)
	}

	input = input.normalized()
	if err := validateUpdateInput(input); err != nil {
		return Request{}, err
	}

	return s.store.Update(ctx, id, input)
}

func validateCreateInput(input CreateInput) error {
	if input.CustomerName == "" {
		return fmt.Errorf("%w: customer_name is required", ErrValidation)
	}
	if input.Contact == "" {
		return fmt.Errorf("%w: contact is required", ErrValidation)
	}
	if input.CargoName == "" {
		return fmt.Errorf("%w: cargo_name is required", ErrValidation)
	}
	if input.DestinationCity == "" {
		return fmt.Errorf("%w: destination_city is required", ErrValidation)
	}
	if input.Comment == "" {
		return fmt.Errorf("%w: comment is required", ErrValidation)
	}
	if !hasPositiveMeasurement(input.WeightKg, input.VolumeM3) {
		return fmt.Errorf("%w: weight_kg or volume_m3 is required", ErrValidation)
	}
	if _, ok := allowedStatuses[input.Status]; !ok {
		return fmt.Errorf("%w: unsupported status", ErrValidation)
	}

	return nil
}

func validateUpdateInput(input UpdateInput) error {
	if input.Status == nil && input.ManagerComment == nil {
		return fmt.Errorf("%w: at least one field is required", ErrValidation)
	}
	if input.Status != nil {
		if _, ok := allowedStatuses[*input.Status]; !ok {
			return fmt.Errorf("%w: unsupported status", ErrValidation)
		}
	}

	return nil
}

func hasPositiveMeasurement(weightKg, volumeM3 *float64) bool {
	if weightKg != nil && *weightKg > 0 {
		return true
	}
	if volumeM3 != nil && *volumeM3 > 0 {
		return true
	}

	return false
}
