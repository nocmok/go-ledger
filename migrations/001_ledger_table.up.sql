create table if not exists ledger (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    metadata jsonb not null default '{}',
    created_at timestamp not null default now()
);
