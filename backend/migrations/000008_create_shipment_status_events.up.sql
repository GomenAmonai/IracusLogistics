-- История смены статусов груза. Пишется при создании груза (начальный статус) и при
-- каждой смене статуса менеджером. Клиент видит таймлайн в WebApp.
create table shipment_status_events (
    id          uuid primary key default gen_random_uuid(),
    shipment_id uuid not null references shipments (id) on delete cascade,
    status      varchar(20) not null
                check (status in ('pending', 'picked_up', 'in_transit', 'customs_clear',
                                  'in_warehouse', 'out_for_delivery', 'delivered', 'cancelled')),
    comment     text,
    -- changed_by nullable + SET NULL: историю сохраняем, даже если менеджера удалили.
    changed_by  uuid references managers (id) on delete set null,
    created_at  timestamptz not null default now()
);

create index idx_shipment_status_events_shipment_id on shipment_status_events (shipment_id);
