create table if not exists ledger_entry
(
    ledger_id uuid not null, 
    transaction_id uuid not null,
    foreign key (ledger_id, transaction_id) references transaction(ledger_id, id),
    id uuid primary key not null,
    account_id uuid not null,
    foreign key (ledger_id, account_id) references account(ledger_id, id),
    amount bigint
);