-- +goose Up

PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS actors (
  Id string PRIMARY KEY,
  Type string NOT NULL,
  Email string,
  PrivateKey BLOB,
  PrivateKeySalt BLOB,
  PublicKey string,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  Properties string NOT NULL DEFAULT (json_object())
);

CREATE INDEX IF NOT EXISTS actors_email ON actors(email);

CREATE TABLE IF NOT EXISTS actor_following (
  Id string PRIMARY KEY,
  ActorId string NOT NULL,
  TargetActorId string NOT NULL,
  TargetActorAccount string NOT NULL,
  State string NOT NULL DEFAULT 'pending',
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  FOREIGN KEY(ActorId)  REFERENCES actors(Id),
  FOREIGN KEY(TargetActorId)  REFERENCES actors(Id)
);

CREATE UNIQUE INDEX IF NOT EXISTS actor_following_actor_id ON actor_following(ActorId, TargetActorId);
CREATE INDEX IF NOT EXISTS actor_following_target_actor_id ON actor_following(TargetActorId);

CREATE TABLE IF NOT EXISTS objects (
  Id string PRIMARY KEY,
  MastodonId string UNIQUE NOT NULL,
  Type string NOT NULL,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  OriginalActorId string,
  OriginalObjectId string UNIQUE,
  ReplyToObjectId string,
  Properties string NOT NULL DEFAULT (json_object()),
  Local INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS inbox_objects (
  Id string PRIMARY KEY,
  ActorId string NOT NULL,
  ObjectId string NOT NULL,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  FOREIGN KEY(ActorId)  REFERENCES actors(Id),
  FOREIGN KEY(ObjectId) REFERENCES objects(Id)
);

CREATE TABLE IF NOT EXISTS outbox_objects (
  Id string PRIMARY KEY,
  ActorId string NOT NULL,
  ObjectId string NOT NULL,
  Target TEXT NOT NULL DEFAULT 'as:Public',
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
  PublishedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  FOREIGN KEY(ActorId)  REFERENCES actors(Id),
  FOREIGN KEY(ObjectId) REFERENCES objects(Id)
);

CREATE TABLE IF NOT EXISTS actor_notifications (
  Id INTEGER PRIMARY KEY AUTOINCREMENT,
  Type string NOT NULL,
  ActorId string NOT NULL,
  FromActorId string NOT NULL,
  ObjectId string,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  FOREIGN KEY(ActorId)  REFERENCES actors(id),
  FOREIGN KEY(FromActorId)  REFERENCES actors(id),
  FOREIGN KEY(ObjectId) REFERENCES objects(id)
);

CREATE INDEX IF NOT EXISTS actor_notifications_actor_id ON actor_notifications(ActorId);

CREATE TABLE IF NOT EXISTS actor_favorites (
  Id string PRIMARY KEY,
  ActorId string NOT NULL,
  ObjectId string NOT NULL,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  FOREIGN KEY(ActorId)  REFERENCES actors(id),
  FOREIGN KEY(ObjectId) REFERENCES objects(id)
);

CREATE INDEX IF NOT EXISTS actor_favorites_actor_id ON actor_favorites(ActorId);
CREATE INDEX IF NOT EXISTS actor_favorites_object_id ON actor_favorites(ObjectId);

CREATE TABLE IF NOT EXISTS actor_reblogs (
  Id string PRIMARY KEY,
  ActorId string NOT NULL,
  ObjectId string NOT NULL,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  FOREIGN KEY(ActorId)  REFERENCES actors(id),
  FOREIGN KEY(ObjectId) REFERENCES objects(id)
);

CREATE INDEX IF NOT EXISTS actor_reblogs_actor_id ON actor_reblogs(ActorId);
CREATE INDEX IF NOT EXISTS actor_reblogs_object_id ON actor_reblogs(ObjectId);

CREATE TABLE IF NOT EXISTS subscriptions (
  Id string PRIMARY KEY,
  ActorId string NOT NULL,
  ClientId string NOT NULL,
  Endpoint string NULL,
  KeyP256dh string NOT NULL,
  KeyAuth string NOT NULL,
  AlertMention INTEGER NOT NULL,
  AlertStatus INTEGER NOT NULL,
  AlertReblog INTEGER NOT NULL,
  AlertFollow INTEGER NOT NULL,
  AlertFollowRequest INTEGER NOT NULL,
  AlertFavorite INTEGER NOT NULL,
  AlertPoll INTEGER NOT NULL,
  AlertUpdate INTEGER NOT NULL,
  AlertAdminSignUp INTEGER NOT NULL,
  AlertAdminReport INTEGER NOT NULL,
  Policy string,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  UNIQUE(ActorId, ClientId)
  FOREIGN KEY(ActorId)  REFERENCES actors(id),
  FOREIGN KEY(ClientId) REFERENCES clients(id)
);

CREATE VIRTUAL TABLE IF NOT EXISTS search_fts USING fts5 (
    Type,
    Name,
    PreferredUsername,
    Status
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS actors_search_fts_insert AFTER INSERT ON actors
BEGIN
    INSERT INTO search_fts (rowid, Type, Name, PreferredUsername)
    VALUES (new.rowid,
            new.type,
            json_extract(new.properties, '$.name'),
            json_extract(new.properties, '$.preferredUsername'));
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS actors_search_fts_delete AFTER DELETE ON actors
BEGIN
    DELETE FROM search_fts WHERE rowid=old.rowid;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS actors_search_fts_update AFTER UPDATE ON actors
BEGIN
    DELETE FROM search_fts WHERE rowid=old.rowid;
    INSERT INTO search_fts (rowid, Type, Name, PreferredUsername)
    VALUES (new.rowid,
            new.type,
            json_extract(new.properties, '$.name'),
            json_extract(new.properties, '$.preferredUsername'));
END;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS actor_replies (
  Id string PRIMARY KEY,
  ActorId string NOT NULL,
  ObjectId string NOT NULL,
  InReplyToObjectId string NOT NULL,
  CreatedAt DATETIME NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),

  FOREIGN KEY(ActorId)  REFERENCES actors(id),
  FOREIGN KEY(ObjectId) REFERENCES objects(id)
  FOREIGN KEY(InReplyToObjectId) REFERENCES objects(id)
);

CREATE INDEX IF NOT EXISTS actor_replies_in_reply_to_object_id ON actor_replies(InReplyToObjectId);

-- +goose Down
DROP TABLE subscriptions;
DROP TABLE actor_favorites;
DROP TABLE actor_following;
DROP TABLE actor_reblogs;
DROP TABLE actor_replies;
DROP TABLE actor_notifications;
DROP TABLE inbox_objects;
DROP TABLE outbox_objects;
DROP TABLE objects;
DROP TABLE actors;