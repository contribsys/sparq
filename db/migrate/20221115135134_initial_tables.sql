-- +goose Up
create table if not exists `accounts` (
  id         integer primary key,
  sfid       string not null, -- snowflake id
  fullname   string not null,
  nick       string not null,
  email      string not null,
  rolemask   integer not null default 1,
  visibility integer not null default 0,
  createdat  timestamp not null default current_timestamp,
  updatedat  timestamp not null default current_timestamp,
  unique (sfid),
  unique (nick)
);
create table if not exists `oauth_clients` (
  Name           string not null,
  ClientId       string not null,
  Secret         string not null,
  RedirectUris   string not null,
  Website        string not null,
  UserId         integer,
  Scopes         string not null default "read",
  CreatedAt      timestamp not null default current_timestamp,
  unique (ClientId),
  foreign key (UserId) references accounts(id) on delete cascade 
);
create table if not exists `oauth_tokens` (
	ClientId            string not null,
	UserId              integer not null, 
	RedirectUri         string not null, 
	Scope               string not null, 
	Code                string not null, 
	CodeChallenge       string,
	CodeCreatedAt        timestamp,
	CodeExpiresIn       integer,
	Access              string, 
	AccessCreatedAt      timestamp,
	AccessExpiresIn     integer,
	Refresh             string, 
	RefreshCreatedAt     timestamp,
	RefreshExpiresIn    integer,
  CreatedAt           timestamp not null default current_timestamp,
  foreign key (UserId) references accounts(id) on delete cascade 
  foreign key (ClientId) references oauth_clients(ClientId) on delete cascade 
);
create index idx_oauth_tokens_code on oauth_tokens(code);
create index idx_oauth_tokens_access on oauth_tokens(access);
create index idx_oauth_tokens_refresh on oauth_tokens(refresh);

create table if not exists `account_profiles` (
	accountid integer not null,
	note      string not null default "",
	avatar    string not null default "/static/default_avatar.png",
	header    string not null default "/static/default_header.jpg",
  foreign key (accountid) references accounts(id) on delete cascade 
);
create table if not exists `account_fields` (
  accountid    integer not null,
  name         string not null,
  value        string not null,
  verifiedat  timestamp,
  foreign key (accountid) references accounts(id) on delete cascade 
);
create table if not exists `account_securities` (
  accountid     integer primary key,
  passwordhash  blob not null,
  publickey     string not null,
  privatekey    string not null,
  foreign key (AccountId) references accounts(id) on delete cascade 
);
create table if not exists `collections` (
  Id integer primary key,
  AccountId integer not null,
  Title string not null,
  Description string not null,
  Visibility integer not null default 0,
  foreign key (AccountId) references accounts(id) on delete cascade 
);
create table if not exists `actors` (
  id string primary key, -- "https://instance.domain/@username"
  accountid integer, -- if this is a local user, this will be non-null
  inbox string, -- "https://instance.domain/@username/inbox"
  sharedinbox string, -- "https://instance.domain/@username"
  foreign key (accountid) references accounts(id) on delete cascade
);
create table if not exists `posts` (
  Id integer primary key,
  URI string not null,
  AuthorId integer not null,
  InReplyToAccountId integer,
  InReplyTo string,
  BoostOfId integer,
  Summary string,
  Content string not null,
  Lang string default 'en',
  Visibility integer not null default 0,
  CollectionId integer,
  PollId integer,
  LastEditAt timestamp,
  DeletedAt timestamp,
  CreatedAt timestamp not null default current_timestamp,
  UpdatedAt timestamp not null default current_timestamp,
  unique (uri),
  foreign key (AuthorId) references Accounts(Id) on delete cascade,
  foreign key (InReplyToAccountId) references Accounts(Id) on delete cascade,
  foreign key (InReplyTo) references Posts(uri) on delete cascade,
  foreign key (BoostOfId) references Posts(id) on delete cascade,
  foreign key (CollectionId) references collections(id) on delete set null
);

-- +goose Down
drop table posts;
drop table actors;
drop table collections;
drop table account_securities;
drop table account_fields;
drop table oauth_clients;
drop table oauth_tokens;
drop table accounts;