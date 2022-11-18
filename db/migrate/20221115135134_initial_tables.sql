-- +goose Up
create table if not exists `users` (
  Id         integer primary key,
  Sfid       string not null, -- snowflake id
  FullName   string not null,
  Nick       string not null,
  Email      string not null,
  RoleMask   integer not null,
  CreatedAt  timestamp not null default current_timestamp,
  UpdatedAt  timestamp not null default current_timestamp
);
create table if not exists `user_attributes` (
  Id         integer primary key,
  UserId     integer not null,
  Name       string not null,
  Value      string not null,
  FOREIGN KEY (UserId) 
      REFERENCES users (id) 
         ON DELETE CASCADE 
);
create table if not exists `user_securities` (
  UserId        integer primary key,
  PasswordHash  string not null,
  PublicKey     string not null,
  PrivateKey    string not null,
  FOREIGN KEY (UserId) 
      REFERENCES users (id) 
         ON DELETE CASCADE 
);
create table if not exists `collections` (
  id integer primary key,
  alias string,
  title string not null,
  description string not null,
  format string,
  privacy tinyint(1) not null,
  owner_id integer NOT NULL,
  FOREIGN KEY (owner_id) 
      REFERENCES users (id) 
         ON DELETE CASCADE 
);
create table if not exists `posts` (
  id integer primary key,
  slug string,
  lang string DEFAULT 'en',
  privacy tinyint(1) NOT NULL,
  owner_id integer not null,
  collection_id integer,
  pindex integer,
  created_at timestamp not null default current_timestamp,
  updated_at timestamp not null default current_timestamp,
  content string not null,
  FOREIGN KEY (owner_id) 
      REFERENCES users (id) 
         ON DELETE CASCADE,
  FOREIGN KEY (collection_id) 
      REFERENCES collections (id) 
         ON DELETE SET NULL
         ON UPDATE NO ACTION
);
create table if not exists `remoteusers` (
  id         integer primary key,
  actor_id string NOT NULL,
  inbox string NOT NULL,
  shared_inbox string NOT NULL
);

-- +goose Down
drop table remoteusers;
drop table posts;
drop table collections;
drop table user_securities;
drop table user_attributes;
drop table users;