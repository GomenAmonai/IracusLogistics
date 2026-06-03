alter table clients
    add column lead_id uuid references leads (id) on delete set null;

create index idx_clients_lead_id on clients (lead_id);
