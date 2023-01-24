-- +goose Up
PRAGMA foreign_keys = ON;

create table if not exists `accounts` (
  Id         integer primary key,
  Sid        string not null,
  FullName   string not null,
  Nick       string not null,
  Email      string not null,
  RoleMask   integer not null default 1,
  Visibility integer not null default 0,
  CreatedAt  timestamp not null default current_timestamp,
  UpdatedAt  timestamp not null default current_timestamp,
  unique (Sid),
  unique (Nick)
);
create table if not exists `oauth_clients` (
  Id             integer primary key,
  ClientId       string not null,
  Name           string not null,
  Secret         string not null,
  RedirectUris   string not null,
  Website        string not null,
  AccountId      integer not null default 0,
  Scopes         string not null default "read",
  CreatedAt      timestamp not null default current_timestamp,
  unique (ClientId),
  foreign key (AccountId) references accounts(Id) on delete cascade 
);
create table if not exists `oauth_tokens` (
	ClientId            string not null,
	AccountId           integer not null, 
	RedirectUri         string not null, 
	Scope               string not null, 
	Code                string not null, 
	CodeChallenge       string,
	CodeCreatedAt       timestamp,
	CodeExpiresIn       integer,
	Access              string, 
	AccessCreatedAt     timestamp,
	AccessExpiresIn     integer,
	Refresh             string, 
	RefreshCreatedAt    timestamp,
	RefreshExpiresIn    integer,
  CreatedAt           timestamp not null default current_timestamp,
  foreign key (AccountId) references accounts(id) on delete cascade 
  foreign key (ClientId) references oauth_clients(ClientId) on delete cascade 
);
create index idx_oauth_tokens_code on oauth_tokens(code);
create index idx_oauth_tokens_access on oauth_tokens(access);
create index idx_oauth_tokens_refresh on oauth_tokens(refresh);

create table if not exists `account_profiles` (
	AccountId integer not null,
	Note      string not null default "",
	Avatar    string not null default "/static/default_avatar.png",
	Header    string not null default "/static/default_header.jpg",
  foreign key (accountid) references accounts(id) on delete cascade 
);
create table if not exists `account_fields` (
  AccountId    integer not null,
  Name         string not null,
  Value        string not null,
  VerifiedAt  timestamp,
  foreign key (accountid) references accounts(id) on delete cascade 
);
create table if not exists `account_securities` (
  AccountId     integer primary key,
  PasswordHash  blob not null,
  PublicKey     string not null,
  PrivateKey    string not null,
  foreign key (AccountId) references accounts(id) on delete cascade 
);
create table if not exists `toots` (
  Sid string not null primary key,
  Uri string not null,
  ActorId integer not null,
  AuthorId integer,
  InReplyToAccountId integer,
  InReplyTo string,
  BoostOfId string,
  Summary string,
  Content string not null,
  Lang string default 'en',
  Visibility integer not null default 0,
  AppId integer,
  PollId integer,
  LastEditAt timestamp,
  DeletedAt timestamp,
  CreatedAt timestamp not null default current_timestamp,
  UpdatedAt timestamp not null default current_timestamp,
  unique (Uri),
  unique (Sid),
  foreign key (AppId) references oauth_clients(Id) on delete set null,
  foreign key (AuthorId) references Accounts(Id) on delete cascade,
  foreign key (InReplyToAccountId) references Accounts(Id) on delete cascade,
  foreign key (InReplyTo) references Toots(uri) on delete cascade,
  foreign key (BoostOfId) references Toots(id) on delete cascade
);
create table if not exists `toot_medias` (
  Id integer primary key,
  Sid string not null default "", -- clients upload media before toot is created
  AccountId integer not null,
  Salt integer not null,
  MimeType string not null default "image/jpeg",
  Path string not null default "/static/undefined.jpg",
  ThumbMimeType string not null default "image/jpeg",
  ThumbPath string not null default "/static/undefined.jpg",
  Meta string default "{}" not null,
  Description string default "",
  Blurhash string default "" not null,
  CreatedAt timestamp not null default current_timestamp,
  foreign key (Sid) references toots(Sid) on delete cascade
  foreign key (AccountId) references accounts(Id) on delete cascade
);
create table if not exists `toot_tags` (
  Sid string not null,
  Tag string not null,
  CreatedAt timestamp not null default current_timestamp,
  foreign key (Sid) references toots(Sid) on delete cascade
);
create index idx_toot_tags_tag on toot_tags(Tag);

-- +goose Down
drop table toot_medias;
drop table toot_tags;
drop table toots;
drop table account_securities;
drop table account_fields;
drop table oauth_clients;
drop table oauth_tokens;
drop table accounts;