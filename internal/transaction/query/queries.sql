-- name: CreateTransaction :one
insert into transaction(ledger_id, currency, status, deadline, metadata)
values ($1, $2, $3, $4, $5)
returning id, ledger_id, currency, status, metadata, error_code, error_message;

-- name: GetTransaction :one
select id, ledger_id, currency, status, metadata, error_code, error_message
from transaction
where id = $2 and ledger_id = $1;

-- name: CreateEntries :copyfrom
insert into transaction_entry (ledger_id, transaction_id, account_id, amount)
values ($1, $2, $3, $4);

-- name: GetEntries :many
select account_id, amount
from transaction_entry
where transaction_id = $2 and ledger_id = $1;