create table messages (
    id          uuid primary key default gen_random_uuid(),
    shipment_id uuid references shipments (id) on delete set null,
    client_id   uuid not null references clients (id) on delete restrict,
    manager_id  uuid references managers (id) on delete set null,
    text        text not null,
    from_role   varchar(10) not null check (from_role in ('client', 'manager')),
    created_at  timestamptz not null default now()
);

create index idx_messages_shipment_id on messages (shipment_id);
create index idx_messages_client_id on messages (client_id);
