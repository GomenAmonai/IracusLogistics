create table leads (
    id         uuid primary key default gen_random_uuid(),
    name       varchar(255) not null,
    phone      varchar(255) not null,
    from_city  varchar(255) not null,
    to_city    varchar(255) not null,
    weight     numeric,
    volume     numeric,
    cargo_type varchar(255),
    comment    text,
    status     varchar(20) not null default 'new'
               check (status in ('new', 'contacted', 'converted', 'rejected')),
    created_at timestamptz not null default now()
);
