create table if not exists account (
    id uuid primary key default gen_random_uuid(),
    ledger_id uuid not null,
    name text not null,
    currency text not null,
    metadata jsonb not null,
    status text not null,
    created_at timestamp not null default now(),
    idempotency_key uuid not null
);

create unique index if not exists account_idempotency_key_idx on account(idempotency_key);