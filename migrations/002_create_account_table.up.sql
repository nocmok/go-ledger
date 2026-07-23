create table if not exists account (
    id uuid primary key,
    ledger_id uuid not null,
    name text not null,
    currency text not null,
    metadata jsonb not null,
    status text not null,
    created_at timestamp with time zone not null default now(),
    idempotency_key uuid not null
);

