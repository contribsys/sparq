-- +goose Up
create table users (
  id          integer primary key,
  username    string not null,
  email       string not null,
  created_at  timestamp not null default current_timestamp,
  updated_at  timestamp not null default current_timestamp,
  deleted_at  timestamp
);

-- +goose Down
drop table users;
