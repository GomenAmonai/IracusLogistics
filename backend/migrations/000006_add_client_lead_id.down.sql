drop index if exists idx_clients_lead_id;

alter table clients
    drop column if exists lead_id;
