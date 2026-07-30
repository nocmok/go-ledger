-- name: CreateAccount :one
insert into account (ledger_id, name, currency, overdraft_allowed, metadata, status)
values ($1, $2, $3, $4, $5, $6)
returning id, ledger_id, name, currency, balance, available_balance, overdraft_allowed, metadata, status;

-- name: GetAccount :one
select id, ledger_id, name, currency, balance, available_balance, overdraft_allowed, metadata, status 
from account
where id = $2 and ledger_id = $1;
