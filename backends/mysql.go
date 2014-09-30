package backends

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var PREPARE_MYSQL = []string{
	fmt.Sprintf(`CREATE TABLE IF NOT EXISTS Users (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		password TEXT NOT NULL,
		state INTEGER DEFAULT %d,
		CONSTRAINT SingleKeys UNIQUE (name) ON CONFLICT ROLLBACK
	);`, STATUS_ACTIVE),
	`CREATE TABLE IF NOT EXISTS Groups (
		gid INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		CONSTRAINT SingleKeys UNIQUE (name) ON CONFLICT ROLLBACK
	);`,
	`CREATE TABLE IF NOT EXISTS UserGroups (
		uid INTEGER NOT NULL REFERENCES Users(uid) ON UPDATE CASCADE,
		gid INTEGER NOT NULL REFERENCES Groups(gid) ON UPDATE CASCADE,
		CONSTRAINT SingleKeys UNIQUE (uid, gid) ON CONFLICT IGNORE
	);`,
	`CREATE TABLE IF NOT EXISTS UserValues (
		uid INTEGER REFERENCES Users(uid) ON UPDATE CASCADE,
		key TEXT NOT NULL,
		value BLOB NOT NULL,
		CONSTRAINT SingleKeys UNIQUE (uid, key) ON CONFLICT REPLACE
	);`,
}

type MysqlBackend struct {
	SqlBackend
}

func NewMysqlBackend(url string) (MysqlBackend, error) {
	var backend MysqlBackend
	db, err := sql.Open("mysql", url)
	if err == nil {
		backend = MysqlBackend{SqlBackend{db: db}}
		return backend, backend.init(PREPARE_MYSQL)
	} else {
		return backend, err
	}
}
