-- name: CreateAccount :one
with inserted as (
    insert into account (ledger_id, name, currency, metadata, status, idempotency_key)
    values ($1, $2, $3, $4, $5, $6)
    on conflict (idempotency_key) do nothing
    returning id, ledger_id, name, currency, metadata, status
)
select id, ledger_id, name, currency, metadata, status from inserted
union all
select id, ledger_id, name, currency, metadata, status from account where idempotency_key = $6;

-- name: GetAccount :one
select id, ledger_id, name, currency, metadata, status 
from account
where id = $1;
