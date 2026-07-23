create table if not exists account (
    id uuid primary key,
    idempotency_key uuid not null,
    metadata jsonb not null,
    status text not null,
    created_at timestamp with time zone default now()
);