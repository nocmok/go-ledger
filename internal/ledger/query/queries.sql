-- name: CreateLedger :one
with inserted as (
    insert into ledger (name, metadata, idempotency_key)
    values ($1, $2, $3)
    on conflict (idempotency_key) do nothing
    returning id, name, metadata
)
select id, name, metadata from inserted
union all
select id, name, metadata from ledger where idempotency_key = $3;

-- name: GetLedger :one
select id, name, metadata from ledger
where id = $1;
