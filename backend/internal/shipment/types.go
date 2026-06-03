package shipment

import (
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusNew                Status = "new"
	StatusNeedsClarification Status = "needs_clarification"
	StatusInCalculation       Status = "in_calculation"
	StatusPriced             Status = "priced"
	StatusOfferSent          Status = "offer_sent"
	StatusWon                Status = "won"
	StatusLost               Status = "lost"
)

var allowedStatuses = map[Status]struct{}{
	StatusNew:                {},
	StatusNeedsClarification: {},
	StatusInCalculation:       {},
	StatusPriced:             {},
	StatusOfferSent:          {},
	StatusWon:                {},
	StatusLost:               {},
}

var ErrValidation = errors.New("shipment validation failed")
var ErrNotFound = errors.New("shipment request not found")

type Request struct {
	ID              string     `json:"id"`
	Status          Status     `json:"status"`
	CustomerName    string     `json:"customer_name"`
	Contact         string     `json:"contact"`
	CompanyName     *string    `json:"company_name,omitempty"`
	CargoName       string     `json:"cargo_name"`
	OriginCity      *string    `json:"origin_city,omitempty"`
	DestinationCity string     `json:"destination_city"`
	WeightKg        *float64   `json:"weight_kg,omitempty"`
	VolumeM3        *float64   `json:"volume_m3,omitempty"`
	BoxesCount      *int       `json:"boxes_count,omitempty"`
	CargoValue      *float64   `json:"cargo_value,omitempty"`
	CargoCurrency   *string    `json:"cargo_currency,omitempty"`
	Comment         string     `json:"comment"`
	ManagerComment  *string    `json:"manager_comment,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateInput struct {
	Status          Status    `json:"status,omitempty"`
	CustomerName    string    `json:"customer_name"`
	Contact         string    `json:"contact"`
	CompanyName     *string   `json:"company_name,omitempty"`
	CargoName       string    `json:"cargo_name"`
	OriginCity      *string   `json:"origin_city,omitempty"`
	DestinationCity string    `json:"destination_city"`
	WeightKg        *float64  `json:"weight_kg,omitempty"`
	VolumeM3        *float64  `json:"volume_m3,omitempty"`
	BoxesCount      *int      `json:"boxes_count,omitempty"`
	CargoValue      *float64  `json:"cargo_value,omitempty"`
	CargoCurrency   *string   `json:"cargo_currency,omitempty"`
	Comment         string    `json:"comment"`
}

type UpdateInput struct {
	Status         *Status `json:"status,omitempty"`
	ManagerComment *string `json:"manager_comment,omitempty"`
}

func (i CreateInput) normalized() CreateInput {
	i.CustomerName = strings.TrimSpace(i.CustomerName)
	i.Contact = strings.TrimSpace(i.Contact)
	i.CargoName = strings.TrimSpace(i.CargoName)
	i.DestinationCity = strings.TrimSpace(i.DestinationCity)
	i.Comment = strings.TrimSpace(i.Comment)
	if i.CompanyName != nil {
		v := strings.TrimSpace(*i.CompanyName)
		if v == "" {
			i.CompanyName = nil
		} else {
			i.CompanyName = &v
		}
	}
	if i.OriginCity != nil {
		v := strings.TrimSpace(*i.OriginCity)
		if v == "" {
			i.OriginCity = nil
		} else {
			i.OriginCity = &v
		}
	}
	if i.CargoCurrency != nil {
		v := strings.TrimSpace(*i.CargoCurrency)
		if v == "" {
			i.CargoCurrency = nil
		} else {
			i.CargoCurrency = &v
		}
	}
	return i
}

func (i UpdateInput) normalized() UpdateInput {
	if i.Status != nil {
		v := Status(strings.TrimSpace(string(*i.Status)))
		i.Status = &v
	}
	if i.ManagerComment != nil {
		v := strings.TrimSpace(*i.ManagerComment)
		if v == "" {
			i.ManagerComment = nil
		} else {
			i.ManagerComment = &v
		}
	}
	return i
}
