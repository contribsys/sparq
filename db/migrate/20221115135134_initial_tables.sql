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
  Id integer primary key,
  ClientId       string not null,
  Name           string not null,
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
  Id integer not null primary key,
  AccountId integer, -- if this is a local user, this will be non-null
  Url string, -- "https://instance.domain/@username"
  Inbox string, -- "https://instance.domain/@username/inbox"
  SharedInbox string, -- "https://instance.domain/@username"
  foreign key (accountid) references accounts(id) on delete cascade
);

create table if not exists `toots` (
  sid string not null primary key,
  uri string not null,
  actorid integer not null,
  authorid integer,
  inreplytoaccountid integer,
  inreplyto string,
  boostofid string,
  summary string,
  content string not null,
  lang string default 'en',
  visibility integer not null default 0,
  collectionid integer,
  appid integer,
  pollid integer,
  lasteditat timestamp,
  deletedat timestamp,
  createdat timestamp not null default current_timestamp,
  updatedat timestamp not null default current_timestamp,
  unique (uri),
  unique (sid),
  foreign key (AppId) references oauth_clients(Id) on delete set null,
  foreign key (ActorId) references Actors(Id) on delete cascade,
  foreign key (AuthorId) references Accounts(Id) on delete cascade,
  foreign key (InReplyToAccountId) references Accounts(Id) on delete cascade,
  foreign key (InReplyTo) references Toots(uri) on delete cascade,
  foreign key (BoostOfId) references Toots(id) on delete cascade,
  foreign key (CollectionId) references collections(id) on delete set null
);

create table if not exists `toot_medias` (
  id integer primary key,
  sid string, -- clients upload media before toot is created
  accountid integer not null,
  mimetype string not null default "image/jpeg",
  uri string not null default "/static/undefined.jpg",
  thumbmimetype string not null default "image/jpeg",
  thumburi string not null default "/static/undefined.jpg",
  meta string default "{}" not null,
  description string default "",
  blurhash string default "" not null,
  height integer,
  width integer,
  duration integer,
  createdat timestamp not null default current_timestamp,
  foreign key (sid) references toots(sid) on delete cascade
  foreign key (accountid) references accounts(id) on delete cascade
);
create table if not exists `toot_tags` (
  Sid string not null,
  Tag string not null,
  CreatedAt timestamp not null default current_timestamp,
  foreign key (sid) references toots(sid) on delete cascade
);
create index idx_toot_tags_tag on toot_tags(tag);

-- +goose Down
drop table toot_medias;
drop table toot_tags;
drop table toots;
drop table actors;
drop table collections;
drop table account_securities;
drop table account_fields;
drop table oauth_clients;
drop table oauth_tokens;
drop table accounts;