create table managers (
    id         uuid primary key default gen_random_uuid(),
    email      varchar(255) not null unique,
    password   varchar(255) not null,
    name       varchar(255) not null,
    created_at timestamptz  not null default now()
);
