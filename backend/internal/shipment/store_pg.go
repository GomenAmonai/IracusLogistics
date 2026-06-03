package shipment

import (
	"context"
	crand "crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStore struct {
	pool *pgxpool.Pool
}

func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

func (s *PGStore) EnsureSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, shipmentSchemaSQL)
	return err
}

func (s *PGStore) Create(ctx context.Context, input CreateInput) (Request, error) {
	now := time.Now().UTC()
	req := Request{
		ID:              newID(),
		Status:          input.Status,
		CustomerName:    input.CustomerName,
		Contact:         input.Contact,
		CompanyName:     input.CompanyName,
		CargoName:       input.CargoName,
		OriginCity:      input.OriginCity,
		DestinationCity: input.DestinationCity,
		WeightKg:        input.WeightKg,
		VolumeM3:        input.VolumeM3,
		BoxesCount:      input.BoxesCount,
		CargoValue:      input.CargoValue,
		CargoCurrency:   input.CargoCurrency,
		Comment:         input.Comment,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	_, err := s.pool.Exec(
		ctx,
		`insert into shipment_requests (
			id, status, customer_name, contact, company_name, cargo_name, origin_city,
			destination_city, weight_kg, volume_m3, boxes_count, cargo_value,
			cargo_currency, comment, created_at, updated_at
		) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		req.ID,
		string(req.Status),
		req.CustomerName,
		req.Contact,
		nullString(req.CompanyName),
		req.CargoName,
		nullString(req.OriginCity),
		req.DestinationCity,
		nullFloat64(req.WeightKg),
		nullFloat64(req.VolumeM3),
		nullInt64(req.BoxesCount),
		nullFloat64(req.CargoValue),
		nullString(req.CargoCurrency),
		req.Comment,
		req.CreatedAt,
		req.UpdatedAt,
	)
	if err != nil {
		return Request{}, err
	}

	return req, nil
}

func (s *PGStore) List(ctx context.Context) ([]Request, error) {
	rows, err := s.pool.Query(ctx, `select
		id, status, customer_name, contact, company_name, cargo_name, origin_city,
		destination_city, weight_kg, volume_m3, boxes_count, cargo_value,
		cargo_currency, comment, manager_comment, created_at, updated_at
		from shipment_requests
		order by created_at desc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Request
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, req)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *PGStore) Get(ctx context.Context, id string) (Request, error) {
	row := s.pool.QueryRow(ctx, `select
		id, status, customer_name, contact, company_name, cargo_name, origin_city,
		destination_city, weight_kg, volume_m3, boxes_count, cargo_value,
		cargo_currency, comment, manager_comment, created_at, updated_at
		from shipment_requests
		where id = $1`, id)

	req, err := scanRequest(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Request{}, ErrNotFound
		}
		return Request{}, err
	}

	return req, nil
}

func (s *PGStore) Update(ctx context.Context, id string, input UpdateInput) (Request, error) {
	now := time.Now().UTC()

	req, err := s.pool.Exec(ctx, `update shipment_requests
		set status = coalesce($2, status),
		    manager_comment = coalesce($3, manager_comment),
		    updated_at = $4
		where id = $1`, id, statusString(input.Status), nullString(input.ManagerComment), now)
	if err != nil {
		return Request{}, err
	}
	if req.RowsAffected() == 0 {
		return Request{}, ErrNotFound
	}

	return s.Get(ctx, id)
}

type requestRowScanner interface {
	Scan(dest ...any) error
}

func scanRequest(row requestRowScanner) (Request, error) {
	var (
		id              string
		status          string
		customerName    string
		contact         string
		companyName     sql.NullString
		cargoName       string
		originCity      sql.NullString
		destinationCity string
		weightKg        sql.NullFloat64
		volumeM3        sql.NullFloat64
		boxesCount      sql.NullInt64
		cargoValue      sql.NullFloat64
		cargoCurrency   sql.NullString
		comment         string
		managerComment  sql.NullString
		createdAt       time.Time
		updatedAt       time.Time
	)

	if err := row.Scan(
		&id,
		&status,
		&customerName,
		&contact,
		&companyName,
		&cargoName,
		&originCity,
		&destinationCity,
		&weightKg,
		&volumeM3,
		&boxesCount,
		&cargoValue,
		&cargoCurrency,
		&comment,
		&managerComment,
		&createdAt,
		&updatedAt,
	); err != nil {
		return Request{}, err
	}

	return Request{
		ID:              id,
		Status:          Status(status),
		CustomerName:    customerName,
		Contact:         contact,
		CompanyName:     nullableString(companyName),
		CargoName:       cargoName,
		OriginCity:      nullableString(originCity),
		DestinationCity: destinationCity,
		WeightKg:        nullableFloat64(weightKg),
		VolumeM3:        nullableFloat64(volumeM3),
		BoxesCount:      nullableInt64(boxesCount),
		CargoValue:      nullableFloat64(cargoValue),
		CargoCurrency:   nullableString(cargoCurrency),
		Comment:         comment,
		ManagerComment:  nullableString(managerComment),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

func newID() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		b[0:4],
		b[4:6],
		b[6:8],
		b[8:10],
		b[10:16],
	)
}

func shipmentStatuses() []string {
	return []string{
		string(StatusNew),
		string(StatusNeedsClarification),
		string(StatusInCalculation),
		string(StatusPriced),
		string(StatusOfferSent),
		string(StatusWon),
		string(StatusLost),
	}
}

const shipmentSchemaSQL = `
create table if not exists shipment_requests (
	id text primary key,
	status text not null default 'new',
	customer_name text not null,
	contact text not null,
	company_name text,
	cargo_name text not null,
	origin_city text,
	destination_city text not null,
	weight_kg double precision,
	volume_m3 double precision,
	boxes_count integer,
	cargo_value double precision,
	cargo_currency text,
	comment text not null,
	manager_comment text,
	created_at timestamptz not null,
	updated_at timestamptz not null,
	constraint shipment_requests_status_check check (status in ('new', 'needs_clarification', 'in_calculation', 'priced', 'offer_sent', 'won', 'lost'))
);

create index if not exists idx_shipment_requests_created_at on shipment_requests (created_at desc);
create index if not exists idx_shipment_requests_status on shipment_requests (status);
`

func statusString(status *Status) any {
	if status == nil {
		return nil
	}

	return string(*status)
}

func nullString(v *string) any {
	if v == nil {
		return nil
	}

	return *v
}

func nullFloat64(v *float64) any {
	if v == nil {
		return nil
	}

	return *v
}

func nullInt64(v *int) any {
	if v == nil {
		return nil
	}

	return *v
}

func nullableString(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}

	value := strings.TrimSpace(v.String)
	if value == "" {
		return nil
	}

	return &value
}

func nullableFloat64(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}

	value := v.Float64
	return &value
}

func nullableInt64(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}

	value := int(v.Int64)
	return &value
}
