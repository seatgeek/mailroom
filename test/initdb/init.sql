create table public.users (
  key varchar(255) primary key,
  preferences jsonb,
  identifiers jsonb,
  emails jsonb,
  created_at timestamp default current_timestamp not null,
  updated_at timestamp default current_timestamp not null,
  deleted_at timestamp null
);

create index idx_users_identifiers on public.users using gin (identifiers);
create index idx_users_emails on public.users using gin (emails);
