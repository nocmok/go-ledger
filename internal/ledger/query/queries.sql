-- name: CreateLedger :one
insert into ledger (name, metadata)
values ($1, $2)
returning id, name, metadata;

-- name: GetLedger :one
select id, name, metadata from ledger
where id = $1;
