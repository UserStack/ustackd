package backends

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var PREPARE_POSTGRES = []string{
	fmt.Sprintf(`CREATE TABLE IF NOT EXISTS Users (
		uid INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		password TEXT NOT NULL,
		state INTEGER DEFAULT %d,
		CONSTRAINT UniqueUserNames UNIQUE (name)
	);`, STATUS_ACTIVE),
	`CREATE TABLE IF NOT EXISTS Groups (
		gid INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		CONSTRAINT UniqueGroupNames UNIQUE (name)
	);`,
	`CREATE TABLE IF NOT EXISTS UserGroups (
		uid INTEGER NOT NULL REFERENCES Users(uid),
		gid INTEGER NOT NULL REFERENCES Groups(gid),
		CONSTRAINT UniqueUidGidPairs UNIQUE (uid, gid)
	);`,
	`CREATE TABLE IF NOT EXISTS UserValues (
		uid INTEGER REFERENCES Users(uid),
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		CONSTRAINT UniqueUidKeyPairs UNIQUE (uid, key)
	);`,
}

type PostgresBackend struct {
	SqlBackend
}

func NewPostgresBackend(url string) (PostgresBackend, error) {
	var backend PostgresBackend
	db, err := sql.Open("postgres", url)
	if err == nil {
		backend = PostgresBackend{SqlBackend{db: db}}
		return backend, backend.init(PREPARE_POSTGRES)
	} else {
		return backend, err
	}
}
