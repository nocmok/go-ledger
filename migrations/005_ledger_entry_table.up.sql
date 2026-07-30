create table if not exists ledger_entry
( 
    transaction_id uuid not null,
    ledger_id uuid not null,
    foreign key (transaction_id, ledger_id) references transaction(id, ledger_id),
    id uuid primary key not null,
    account_id uuid not null,
    foreign key (account_id, ledger_id) references account(id, ledger_id),
    amount bigint
);