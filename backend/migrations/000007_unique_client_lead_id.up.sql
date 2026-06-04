-- Один лид → не более одного клиента. Partial unique: NULL-значения lead_id (клиент
-- пришёл не через форму) не конфликтуют между собой, проверяется только заполненный lead_id.
-- Решение по кардинальности конверсии Lead→Client: 1:1 (см. docs/tech-debt.md #14).
create unique index uq_clients_lead_id on clients (lead_id) where lead_id is not null;
