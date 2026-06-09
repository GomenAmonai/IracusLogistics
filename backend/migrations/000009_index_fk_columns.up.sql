-- Индексы на FK-колонки с ON DELETE SET NULL: без индекса удаление родительской строки
-- (менеджера) вынуждает seq scan дочерней таблицы и берёт лишние блокировки. Отдельной
-- миграцией — для независимого отката (см. database.md).
create index idx_messages_manager_id on messages (manager_id);
create index idx_shipment_status_events_changed_by on shipment_status_events (changed_by);
