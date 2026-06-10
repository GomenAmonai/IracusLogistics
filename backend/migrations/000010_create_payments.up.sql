-- Платежи по грузу. Бизнес мультиканальный: один груз может закрываться несколькими
-- платежами разными способами, поэтому платёж — отдельная таблица, а не поля на грузе.
-- Каналы ратифицированы 2026-06-10: безнал по счёту, карта/СБП, наличные, крипта.
create table payments (
    id          uuid primary key default gen_random_uuid(),
    -- restrict: денежные записи не должны исчезать каскадом — груз с платежами не удалить.
    shipment_id uuid not null references shipments (id) on delete restrict,
    amount      numeric not null check (amount > 0),
    currency    varchar(3) not null,
    channel     varchar(20) not null
                check (channel in ('bank_transfer', 'card_sbp', 'cash', 'crypto')),
    status      varchar(20) not null default 'pending'
                check (status in ('pending', 'confirmed', 'refunded')),
    comment     text,
    -- created_by nullable + SET NULL: платёж сохраняем, даже если менеджера удалили.
    created_by  uuid references managers (id) on delete set null,
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now()
);

create index idx_payments_shipment_id on payments (shipment_id);
create index idx_payments_created_by on payments (created_by);
