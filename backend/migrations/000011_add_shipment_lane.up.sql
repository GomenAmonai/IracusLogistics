-- Полоса доставки: карго / белый импорт / выкуп. Ратифицировано 2026-06-10 как метка на
-- грузе, не раздельные потоки. Существующие грузы — cargo (исторически все были карго).
alter table shipments
    add column lane varchar(10) not null default 'cargo'
    check (lane in ('cargo', 'white', 'buyout'));
