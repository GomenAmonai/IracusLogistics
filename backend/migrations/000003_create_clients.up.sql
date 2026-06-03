create table clients (
    id          uuid primary key default gen_random_uuid(),
    telegram_id bigint not null unique,
    username    varchar(255),
    name        varchar(255) not null,
    phone       varchar(255),
    created_at  timestamptz not null default now()
);
