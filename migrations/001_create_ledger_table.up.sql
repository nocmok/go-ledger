create table if not exists ledger (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    metadata jsonb not null,
    created_at timestamp not null default now(),
    idempotency_key uuid not null
);

create unique index if not exists ledger_idempotency_key_idx on ledger(idempotency_key);