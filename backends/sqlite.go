package backends

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SqliteBackend struct {
	db             *sql.DB
	createUserStmt *sql.Stmt
}

func NewSqliteBackend(url string) (SqliteBackend, error) {
	var backend SqliteBackend
	db, err := sql.Open("sqlite3", url)
	if err == nil {
		backend = SqliteBackend{db: db}
		return backend, backend.init()
	} else {
		return backend, err
	}
}

func (backend *SqliteBackend) init() error {
	var err error
	// initialize all tables
	_, err = backend.db.Exec(`CREATE TABLE IF NOT EXISTS Users (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR,
		password VARCHAR,
		CONSTRAINT SingleKeys UNIQUE (email) ON CONFLICT FAIL
	)`)
	if err != nil {
		return err
	}
	_, err = backend.db.Exec(`CREATE TABLE IF NOT EXISTS Groups (
		gid INTEGER PRIMARY KEY AUTOINCREMENT,
		name VARCHAR,
		CONSTRAINT SingleKeys UNIQUE (name) ON CONFLICT FAIL
	)`)
	if err != nil {
		return err
	}
	_, err = backend.db.Exec(`CREATE TABLE IF NOT EXISTS UserGroups (
		uid INTEGER, gid INTEGER,
		CONSTRAINT SingleKeys UNIQUE (uid, gid) ON CONFLICT IGNORE
	)`)
	if err != nil {
		return err
	}
	_, err = backend.db.Exec(`CREATE TABLE IF NOT EXISTS UserValues (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		key VARCHAR,
		value BLOB,
		CONSTRAINT SingleKeys UNIQUE (uid, key) ON CONFLICT REPLACE
	)`)
	if err != nil {
		return err
	}
	backend.createUserStmt, err = backend.db.Prepare(`INSERT INTO Users (
		email, password
	) VALUES (
		?, ?
	)`)
	if err != nil {
		return err
	}
	return nil
}

func (backend *SqliteBackend) CreateUser(email string, password string) (int, *Error) {
	r, err := backend.createUserStmt.Exec(email, password)
	if email == "" || password == "" {
		return 0, &Error{"EINVAL", "Username and password can't be blank"}
	}
	if err == nil {
		var uid int64
		uid, err = r.LastInsertId()
		if err == nil {
			return int(uid), nil
		} else {
			return 0, &Error{"EFAULT", err.Error()}
		}
	} else {
		return 0, &Error{"EEXIST", err.Error()}
	}
}

func (backend *SqliteBackend) DisableUser(emailuid string) *Error {
	return nil
}

func (backend *SqliteBackend) EnableUser(emailuid string) *Error {
	return nil
}

func (backend *SqliteBackend) SetUserData(emailuid string, key string, value string) *Error {
	return nil
}

func (backend *SqliteBackend) GetUserData(emailuid string, key string) *Error {
	return nil
}

func (backend *SqliteBackend) LoginUser(email string, password string) (int, *Error) {
	return 0, nil
}

func (backend *SqliteBackend) ChangeUserPassword(emailuid string, password string, newpassword string) *Error {
	return nil
}

func (backend *SqliteBackend) ChangeUserEmail(emailuid string, password string, newemail string) *Error {
	return nil
}

func (backend *SqliteBackend) UserGroups(emailuid string) ([]Group, *Error) {
	return nil, nil
}

func (backend *SqliteBackend) DeleteUser(emailuid string) *Error {
	return nil
}

func (backend *SqliteBackend) Users() ([]User, *Error) {
	return nil, nil
}

func (backend *SqliteBackend) Group(name string) (int, *Error) {
	return 0, nil
}

func (backend *SqliteBackend) AddUserToGroup(emailuid string, groupgid string) *Error {
	return nil
}

func (backend *SqliteBackend) RemoveUserFromGroup(emailuid string, groupgid string) *Error {
	return nil
}

func (backend *SqliteBackend) DeleteGroup(groupgid string) *Error {
	return nil
}

func (backend *SqliteBackend) Groups() ([]Group, *Error) {
	return nil, nil
}

func (backend *SqliteBackend) GroupUsers(groupgid string) ([]User, *Error) {
	return nil, nil
}
