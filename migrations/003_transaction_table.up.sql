create table if not exists transaction
(
    ledger_id uuid not null references ledger(id), 
    id uuid not null default gen_random_uuid(),
    primary key (ledger_id, id),
    currency text not null,
    status text not null,
    deadline timestamptz not null,
    metadata jsonb not null default '{}',
    error jsonb,
    created_at timestamptz not null default now(),
    executed_at timestamptz
);