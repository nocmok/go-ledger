create table if not exists account (
    ledger_id uuid not null references ledger(id), 
    id uuid not null default gen_random_uuid(),
    primary key (ledger_id, id),
    name text not null,
    currency text not null,
    balance bigint not null default 0,
    available_balance bigint not null default 0,
    overdraft_allowed bool not null default false,
    metadata jsonb not null default '{}',
    status text not null,
    created_at timestamp not null default now()
);