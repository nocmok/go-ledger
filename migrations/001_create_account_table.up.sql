create table if not exists account (
    id uuid primary key,
    metadata jsonb not null,
    status text not null,
    created_at timestamp with time zone default now()
);