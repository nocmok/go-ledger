create table if not exists transaction
( 
    id uuid not null default gen_random_uuid(),
    ledger_id uuid not null references ledger(id),
    primary key (id, ledger_id),
    currency text not null,
    status text not null,
    deadline timestamptz not null,
    metadata jsonb not null default '{}',
    error_code text,
    error_message text,
    created_at timestamptz not null default now(),
    executed_at timestamptz
);

create table if not exists transaction_entry
(
    transaction_id uuid not null,
    ledger_id uuid not null, 
    foreign key (transaction_id, ledger_id) references transaction(id, ledger_id),
    account_id uuid not null,
    foreign key (account_id, ledger_id) references account(id, ledger_id),
    amount bigint
);

create index if not exists transaction_entry_transaction_id_idx on transaction_entry(transaction_id);