create table shipments (
    id             uuid primary key default gen_random_uuid(),
    client_id      uuid not null references clients (id) on delete restrict,
    manager_id     uuid not null references managers (id) on delete restrict,
    tracking_key   varchar(64) not null unique,
    status         varchar(20) not null default 'pending'
                   check (status in ('pending', 'picked_up', 'in_transit', 'customs_clear',
                                      'in_warehouse', 'out_for_delivery', 'delivered', 'cancelled')),
    status_comment text,
    weight         numeric,
    volume         numeric,
    from_city      varchar(255),
    to_city        varchar(255),
    price          numeric,
    currency       varchar(3) not null default 'USD',
    estimated_at   timestamptz,
    delivered_at   timestamptz,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now()
);

create index idx_shipments_client_id on shipments (client_id);
create index idx_shipments_manager_id on shipments (manager_id);
