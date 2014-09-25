package backends

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

const (
	STATUS_ACTIVE   = 1
	STATUS_INACTIVE = 0
)

type SqliteBackend struct {
	db                 *sql.DB
	createUserStmt     *sql.Stmt
	usersStmt          *sql.Stmt
	deleteUserStmt     *sql.Stmt
	loginUserStmt      *sql.Stmt
	setUserStateStmt   *sql.Stmt
	uidForEmailUidStmt *sql.Stmt
	setUserDataStmt    *sql.Stmt
	getUserDataStmt    *sql.Stmt
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
	_, err = backend.db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS Users (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		email VARCHAR,
		password VARCHAR,
		state INTEGER DEFAULT %d,
		CONSTRAINT SingleKeys UNIQUE (email) ON CONFLICT FAIL
	)`, STATUS_ACTIVE))
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
	backend.usersStmt, err = backend.db.Prepare(`SELECT email, uid FROM Users`)
	if err != nil {
		return err
	}
	backend.deleteUserStmt, err = backend.db.Prepare(`DELETE FROM Users
		WHERE uid = ? OR email = ?`)
	if err != nil {
		return err
	}
	backend.loginUserStmt, err = backend.db.Prepare(fmt.Sprintf(
		"SELECT uid FROM Users WHERE email = ? AND password = ? AND state = %d",
		STATUS_ACTIVE))
	if err != nil {
		return err
	}
	backend.setUserStateStmt, err = backend.db.Prepare(`UPDATE Users
		SET state = ? WHERE email = ? OR uid = ?`)
	if err != nil {
		return err
	}
	backend.uidForEmailUidStmt, err = backend.db.Prepare(`SELECT uid FROM Users
	 	WHERE email = ? OR uid = ?`)
	if err != nil {
		return err
	}
	backend.setUserDataStmt, err = backend.db.Prepare(`INSERT INTO UserValues (
		uid, key, value
	) VALUES (
		?, ?, ?
	)`)
	if err != nil {
		return err
	}
	backend.getUserDataStmt, err = backend.db.Prepare(`SELECT value FROM UserValues
	 	WHERE uid = ? AND key = ?`)
	if err != nil {
		return err
	}
	return nil
}

func (backend *SqliteBackend) CreateUser(email string, password string) (int64, *Error) {
	if email == "" || password == "" {
		return 0, &Error{"EINVAL", "User email and password can't be blank"}
	}
	result, err := backend.createUserStmt.Exec(email, password)
	if err != nil {
		return 0, &Error{"EEXIST", err.Error()}
	}
	var uid int64
	uid, err = result.LastInsertId()
	if err == nil {
		return uid, nil
	} else {
		return 0, &Error{"EFAULT", err.Error()}
	}
}

func (backend *SqliteBackend) DisableUser(emailuid string) *Error {
	return backend.setUserState(STATUS_INACTIVE, emailuid)
}

func (backend *SqliteBackend) EnableUser(emailuid string) *Error {
	return backend.setUserState(STATUS_ACTIVE, emailuid)
}

func (backend *SqliteBackend) SetUserData(emailuid string, key string, value string) *Error {
	if emailuid == "" || key == "" || value == "" {
		return &Error{"EINVAL", "Email/uid, key and value can't be blank"}
	}
	uid, err := backend.getUidForEmailUid(emailuid)
	if err != nil {
		return err
	}
	_, serr := backend.setUserDataStmt.Exec(uid, key, value)
	if serr != nil {
		return &Error{"EFAULT", serr.Error()}
	}
	return nil
}

func (backend *SqliteBackend) GetUserData(emailuid string, key string) (string, *Error) {
	if emailuid == "" || key == "" {
		return "", &Error{"EINVAL", "Email/uid, key and value can't be blank"}
	}
	uid, err := backend.getUidForEmailUid(emailuid)
	if err != nil {
		return "", err
	}
	rows, gerr := backend.getUserDataStmt.Query(uid, key)
	if gerr != nil {
		return "", &Error{"EFAULT", gerr.Error()}
	}
	if rows.Next() {
		var value string
		serr := rows.Scan(&value)
		if serr != nil {
			return "", &Error{"EFAULT", serr.Error()}
		}
		return value, nil
	}
	return "", &Error{"ENOENT", "Key unknown"}
}

func (backend *SqliteBackend) LoginUser(email string, password string) (int64, *Error) {
	if email == "" || password == "" {
		return 0, &Error{"EINVAL", "Username and password can't be blank"}
	}
	rows, err := backend.loginUserStmt.Query(email, password)
	defer rows.Close()
	if err != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	if !rows.Next() {
		return 0, &Error{"ENOENT", "Email unknown"}
	}
	var uid int64
	serr := rows.Scan(&uid)
	if serr != nil {
		return 0, &Error{"EFAULT", serr.Error()}
	}
	return uid, nil
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
	if emailuid == "" {
		return &Error{"EINVAL", "Email or uid has to be passed"}
	}

	result, err := backend.deleteUserStmt.Exec(emailuid, emailuid)
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}
	count, err := result.RowsAffected()
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}

	if count < 1 {
		return &Error{"ENOENT", "Email or uid unknown"}
	}
	return nil
}

func (backend *SqliteBackend) Users() ([]User, *Error) {
	var users []User
	rows, err := backend.usersStmt.Query()
	defer rows.Close()
	if err != nil {
		return nil, &Error{"EFAULT", err.Error()}
	}
	for rows.Next() {
		var uid int64
		var email string
		err = rows.Scan(&email, &uid)
		if err != nil {
			return nil, &Error{"EFAULT", err.Error()}
		}
		users = append(users, User{uid, email})
	}
	return users, nil
}

func (backend *SqliteBackend) Group(name string) (int64, *Error) {
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

func (backend *SqliteBackend) Close() {
	backend.db.Close()
}

func (backend *SqliteBackend) getUidForEmailUid(emailuid string) (int64, *Error) {
	rows, err := backend.uidForEmailUidStmt.Query(emailuid, emailuid)
	defer rows.Close()
	if err != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	if !rows.Next() {
		return 0, &Error{"ENOENT", "Email unknown"}
	}
	var uid int64
	serr := rows.Scan(&uid)
	if serr != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	return uid, nil
}

func (backend *SqliteBackend) setUserState(state int, emailuid string) *Error {
	if emailuid == "" {
		return &Error{"EINVAL", "User email or uid must be given"}
	}
	result, err := backend.setUserStateStmt.Exec(state, emailuid, emailuid)
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}
	n, aerr := result.RowsAffected()
	if aerr != nil {
		return &Error{"EFAULT", err.Error()}
	}
	if n == 0 {
		return &Error{"ENOENT", "User email"}
	}
	return nil
}
