-- +goose Up
create table if not exists `users` (
  Id         integer primary key,
  Sfid       string not null, -- snowflake id
  FullName   string not null,
  Nick       string not null,
  Email      string not null,
  RoleMask   integer not null,
  CreatedAt  timestamp not null default current_timestamp,
  UpdatedAt  timestamp not null default current_timestamp,
  unique (sfid),
  unique (nick)
);
create table if not exists `user_attributes` (
  Id         integer primary key,
  UserId     integer not null,
  Name       string not null,
  Value      string not null,
  foreign key (UserId) references users(id) on delete cascade 
);
create table if not exists `user_securities` (
  UserId        integer primary key,
  PasswordHash  string not null,
  PublicKey     string not null,
  PrivateKey    string not null,
  foreign key (UserId) references users(id) on delete cascade 
);
create table if not exists `collections` (
  Id integer primary key,
  UserId integer not null,
  Title string not null,
  Description string not null,
  Visibility integer not null default 0,
  foreign key (UserId) references users(id) on delete cascade 
);
create table if not exists `actors` (
  Id string primary key, -- "https://instance.domain/@username"
  UserId integer, -- if this is a local user, this will be non-null
  Inbox string, -- "https://instance.domain/@username/inbox"
  SharedInbox string, -- "https://instance.domain/@username"
  foreign key (UserId) references users(id) on delete cascade
);
create table if not exists `posts` (
  Id integer primary key,
  URI string not null,
  AuthorId string not null,
  InReplyTo string,
  Summary string,
  Content string not null,
  Lang string default 'en',
  Visibility integer not null default 0,
  CollectionId integer,
  CreatedAt timestamp not null default current_timestamp,
  UpdatedAt timestamp not null default current_timestamp,
  unique (URI),
  foreign key (AuthorId) references Actors(Id) on delete cascade,
  -- foreign key (InReplyTo) references Posts(id) on delete cascade,
  foreign key (CollectionId) references collections(id) on delete set null
);

-- +goose Down
drop table posts;
drop table actors;
drop table collections;
drop table user_securities;
drop table user_attributes;
drop table users;