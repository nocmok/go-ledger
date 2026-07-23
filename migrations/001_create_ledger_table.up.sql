create table if not exists ledger (
    id uuid primary key,
    name text not null,
    metadata jsonb not null,
    created_at timestamp with time zone not null default now(),
    idempotency_key uuid not null
);