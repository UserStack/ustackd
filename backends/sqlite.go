package backends

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

const (
	STATUS_ACTIVE          = 1
	STATUS_INACTIVE        = 0
	CREATE_USER_TABLE_STMT = `CREATE TABLE IF NOT EXISTS Users (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		password TEXT NOT NULL,
		state INTEGER DEFAULT %d,
		CONSTRAINT SingleKeys UNIQUE (name) ON CONFLICT ROLLBACK
	);`
	CREATE_GROUP_TABLE_STMT = `CREATE TABLE IF NOT EXISTS Groups (
		gid INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		CONSTRAINT SingleKeys UNIQUE (name) ON CONFLICT ROLLBACK
	);`
	CREATE_USER_GROUP_TABLE_STMT = `CREATE TABLE IF NOT EXISTS UserGroups (
		uid INTEGER NOT NULL REFERENCES Users(uid) ON UPDATE CASCADE,
		gid INTEGER NOT NULL REFERENCES Groups(gid) ON UPDATE CASCADE,
		CONSTRAINT SingleKeys UNIQUE (uid, gid) ON CONFLICT IGNORE
	);`
	CREATE_USER_VALUES_TABLE_STMT = `CREATE TABLE IF NOT EXISTS UserValues (
		uid INTEGER REFERENCES Users(uid) ON UPDATE CASCADE,
		key TEXT NOT NULL,
		value BLOB NOT NULL,
		CONSTRAINT SingleKeys UNIQUE (uid, key) ON CONFLICT REPLACE
	);`
	CREATE_USER_STMT = `INSERT INTO Users (
		name, password
	) VALUES (
		?, ?
	);`
	USERS_STMT       = `SELECT name, uid, state FROM Users`
	DELETE_USER_STMT = `DELETE FROM Users WHERE uid = ? OR name = ?;`
	LOGIN_USER_STMT  = `SELECT uid FROM Users
		WHERE name = ? AND password = ? AND state = %d;`
	SET_USER_STATE_STMT = `UPDATE Users
		SET state = ? WHERE name = ? OR uid = ?;`
	UID_FOR_NAME_UID_STMT = `SELECT uid FROM Users
		WHERE name = ? OR uid = ?;`
	SET_USER_DATA_STMT = `INSERT INTO UserValues (
		uid, key, value
	) VALUES (
		?, ?, ?
	);`
	GET_USER_DATA_STMT = `SELECT value FROM UserValues
		WHERE uid = ? AND key = ?;`
	CHANGE_USER_PASSWD_STMT = `UPDATE Users SET password = ?
		WHERE uid = ? AND password = ?`
	CHANGE_USER_NAME_STMT = `UPDATE Users SET name = ?
		WHERE uid = ? AND password = ?`
	CREATE_GROUP_STMT = `INSERT INTO Groups (name) VALUES (?)`
	GROUPS_STMT       = `SELECT name, gid FROM Groups`
	DELETE_GROUP_STMT = `DELETE FROM Groups WHERE gid = ? OR name = ?;`
)

var PREPARE = []string{
	`PRAGMA encoding = "UTF-8";`,
	"PRAGMA foreign_keys = ON;",
	"PRAGMA journal_mode;",
	"PRAGMA integrity_check;",
	"PRAGMA busy_timeout = 60000;",
	"PRAGMA auto_vacuum = INCREMENTAL;",
}

type SqliteBackend struct {
	db                     *sql.DB
	createUserStmt         *sql.Stmt
	usersStmt              *sql.Stmt
	deleteUserStmt         *sql.Stmt
	loginUserStmt          *sql.Stmt
	setUserStateStmt       *sql.Stmt
	uidForNameUidStmt      *sql.Stmt
	setUserDataStmt        *sql.Stmt
	getUserDataStmt        *sql.Stmt
	changeUserPasswordStmt *sql.Stmt
	changeUserNameStmt     *sql.Stmt
	createGroupStmt        *sql.Stmt
	groupsStmt             *sql.Stmt
	deleteGroupStmt        *sql.Stmt
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
	// set the default encoding, enable foreign keys, enable journal mode,
	// check the integrity, set timeout to 60 sec and enable the auto vacuum
	for _, stmt := range PREPARE {
		_, err = backend.db.Exec(stmt)
		if err != nil {
			return err
		}
	}
	// initialize all tables
	_, err = backend.db.Exec(fmt.Sprintf(CREATE_USER_TABLE_STMT, STATUS_ACTIVE))
	if err != nil {
		return err
	}
	_, err = backend.db.Exec(CREATE_GROUP_TABLE_STMT)
	if err != nil {
		return err
	}
	_, err = backend.db.Exec(CREATE_USER_GROUP_TABLE_STMT)
	if err != nil {
		return err
	}
	_, err = backend.db.Exec(CREATE_USER_VALUES_TABLE_STMT)
	if err != nil {
		return err
	}
	backend.createUserStmt, err = backend.db.Prepare(CREATE_USER_STMT)
	if err != nil {
		return err
	}
	backend.usersStmt, err = backend.db.Prepare(USERS_STMT)
	if err != nil {
		return err
	}
	backend.deleteUserStmt, err = backend.db.Prepare(DELETE_USER_STMT)
	if err != nil {
		return err
	}
	backend.loginUserStmt, err = backend.db.Prepare(fmt.Sprintf(
		LOGIN_USER_STMT, STATUS_ACTIVE))
	if err != nil {
		return err
	}
	backend.setUserStateStmt, err = backend.db.Prepare(SET_USER_STATE_STMT)
	if err != nil {
		return err
	}
	backend.uidForNameUidStmt, err = backend.db.Prepare(UID_FOR_NAME_UID_STMT)
	if err != nil {
		return err
	}
	backend.setUserDataStmt, err = backend.db.Prepare(SET_USER_DATA_STMT)
	if err != nil {
		return err
	}
	backend.getUserDataStmt, err = backend.db.Prepare(GET_USER_DATA_STMT)
	if err != nil {
		return err
	}
	backend.changeUserPasswordStmt, err = backend.db.Prepare(CHANGE_USER_PASSWD_STMT)
	if err != nil {
		return err
	}
	backend.changeUserNameStmt, err = backend.db.Prepare(CHANGE_USER_NAME_STMT)
	if err != nil {
		return err
	}
	backend.createGroupStmt, err = backend.db.Prepare(CREATE_GROUP_STMT)
	if err != nil {
		return err
	}
	backend.groupsStmt, err = backend.db.Prepare(GROUPS_STMT)
	if err != nil {
		return err
	}
	backend.deleteGroupStmt, err = backend.db.Prepare(DELETE_GROUP_STMT)
	if err != nil {
		return err
	}
	return nil
}

func (backend *SqliteBackend) CreateUser(name string, password string) (int64, *Error) {
	if name == "" || password == "" {
		return 0, &Error{"EINVAL", "User name and password can't be blank"}
	}
	result, err := backend.createUserStmt.Exec(name, password)
	if err != nil {
		// after error occured, the statement is broken, we need to recreate it
		backend.createUserStmt.Close()
		backend.createUserStmt, _ = backend.db.Prepare(CREATE_USER_STMT)
		return 0, &Error{"EEXIST", err.Error()}
	}
	var uid int64
	uid, err = result.LastInsertId()
	if err == nil {
		return uid, nil
	} else {
		fmt.Printf("  Err %s\n", err)
		return 0, &Error{"EFAULT", err.Error()}
	}
}

func (backend *SqliteBackend) DisableUser(nameuid string) *Error {
	return backend.setUserState(STATUS_INACTIVE, nameuid)
}

func (backend *SqliteBackend) EnableUser(nameuid string) *Error {
	return backend.setUserState(STATUS_ACTIVE, nameuid)
}

func (backend *SqliteBackend) SetUserData(nameuid string, key string, value string) *Error {
	if nameuid == "" || key == "" || value == "" {
		return &Error{"EINVAL", "Name/uid, key and value can't be blank"}
	}
	uid, err := backend.getUidForNameUid(nameuid)
	if err != nil {
		return err
	}
	_, serr := backend.setUserDataStmt.Exec(uid, key, value)
	if serr != nil {
		return &Error{"EFAULT", serr.Error()}
	}
	return nil
}

func (backend *SqliteBackend) GetUserData(nameuid string, key string) (string, *Error) {
	if nameuid == "" || key == "" {
		return "", &Error{"EINVAL", "Name/uid, key and value can't be blank"}
	}
	uid, err := backend.getUidForNameUid(nameuid)
	if err != nil {
		return "", err
	}
	rows, gerr := backend.getUserDataStmt.Query(uid, key)
	defer rows.Close()
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

func (backend *SqliteBackend) LoginUser(name string, password string) (int64, *Error) {
	if name == "" || password == "" {
		return 0, &Error{"EINVAL", "Username and password can't be blank"}
	}
	rows, err := backend.loginUserStmt.Query(name, password)
	defer rows.Close()
	if err != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	if !rows.Next() {
		return 0, &Error{"ENOENT", "Name unknown"}
	}
	var uid int64
	serr := rows.Scan(&uid)
	if serr != nil {
		return 0, &Error{"EFAULT", serr.Error()}
	}
	return uid, nil
}

func (backend *SqliteBackend) ChangeUserPassword(nameuid string, password string, newpassword string) *Error {
	if nameuid == "" || password == "" || newpassword == "" {
		return &Error{"EINVAL", "nameuid and passwords can't be blank"}
	}
	uid, err := backend.getUidForNameUid(nameuid)
	if err != nil {
		return err
	}
	if uid <= 0 {
		return &Error{"ENOENT", "Password didn't match"}
	}
	result, serr := backend.changeUserPasswordStmt.Exec(newpassword, uid, password)
	if serr != nil {
		return &Error{"EFAULT", serr.Error()}
	}
	count, rerr := result.RowsAffected()
	if rerr != nil {
		return &Error{"EFAULT", rerr.Error()}
	}
	if count < 1 {
		return &Error{"ENOENT", "Password didn't match"}
	}
	return nil
}

func (backend *SqliteBackend) ChangeUserName(nameuid string, password string, newname string) *Error {
	if nameuid == "" || password == "" || newname == "" {
		return &Error{"EINVAL", "nameuid, password and new name can't be blank"}
	}
	uid, err := backend.getUidForNameUid(nameuid)
	if err != nil {
		return err
	}
	if uid <= 0 {
		return &Error{"ENOENT", "Password didn't match"}
	}
	result, serr := backend.changeUserNameStmt.Exec(newname, uid, password)
	if serr != nil {
		return &Error{"EFAULT", serr.Error()}
	}
	count, rerr := result.RowsAffected()
	if rerr != nil {
		return &Error{"EFAULT", rerr.Error()}
	}
	if count < 1 {
		return &Error{"ENOENT", "Password didn't match"}
	}
	return nil
}

func (backend *SqliteBackend) UserGroups(nameuid string) ([]Group, *Error) {
	return nil, nil
}

func (backend *SqliteBackend) DeleteUser(nameuid string) *Error {
	if nameuid == "" {
		return &Error{"EINVAL", "Name or uid has to be passed"}
	}

	result, err := backend.deleteUserStmt.Exec(nameuid, nameuid)
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}
	count, err := result.RowsAffected()
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}

	if count < 1 {
		return &Error{"ENOENT", "Name or uid unknown"}
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
		var name string
		var state int
		err = rows.Scan(&name, &uid, &state)
		if err != nil {
			return nil, &Error{"EFAULT", err.Error()}
		}
		users = append(users, User{uid, name, state == STATUS_ACTIVE})
	}
	return users, nil
}

func (backend *SqliteBackend) CreateGroup(name string) (int64, *Error) {
	if name == "" {
		return 0, &Error{"EINVAL", "Invalid group name"}
	}
	result, err := backend.createGroupStmt.Exec(name)
	if err != nil {
		// after error occured, the statement is broken, we need to recreate it
		backend.createGroupStmt.Close()
		backend.createGroupStmt, _ = backend.db.Prepare(CREATE_GROUP_STMT)
		return 0, &Error{"EEXIST", err.Error()}
	}
	var uid int64
	uid, err = result.LastInsertId()
	if err == nil {
		return uid, nil
	} else {
		fmt.Printf("  Err %s\n", err)
		return 0, &Error{"EFAULT", err.Error()}
	}
}

func (backend *SqliteBackend) AddUserToGroup(nameuid string, groupgid string) *Error {
	return nil
}

func (backend *SqliteBackend) RemoveUserFromGroup(nameuid string, groupgid string) *Error {
	return nil
}

func (backend *SqliteBackend) DeleteGroup(groupgid string) *Error {
	if groupgid == "" {
		return &Error{"EINVAL", "Name or gid has to be passed"}
	}

	result, err := backend.deleteGroupStmt.Exec(groupgid, groupgid)
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}
	count, err := result.RowsAffected()
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}

	if count < 1 {
		return &Error{"ENOENT", "Name or gid unknown"}
	}
	return nil
}

func (backend *SqliteBackend) Groups() ([]Group, *Error) {
	var groups []Group
	rows, err := backend.groupsStmt.Query()
	defer rows.Close()
	if err != nil {
		return nil, &Error{"EFAULT", err.Error()}
	}
	for rows.Next() {
		var gid int64
		var name string
		err = rows.Scan(&name, &gid)
		if err != nil {
			return nil, &Error{"EFAULT", err.Error()}
		}
		groups = append(groups, Group{gid, name})
	}
	return groups, nil
}

func (backend *SqliteBackend) GroupUsers(groupgid string) ([]User, *Error) {
	return nil, nil
}

func (backend *SqliteBackend) Close() {
	backend.db.Close()
}

func (backend *SqliteBackend) getUidForNameUid(nameuid string) (int64, *Error) {
	rows, err := backend.uidForNameUidStmt.Query(nameuid, nameuid)
	defer rows.Close()
	if err != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	if !rows.Next() {
		return 0, &Error{"ENOENT", "Name unknown"}
	}
	var uid int64
	serr := rows.Scan(&uid)
	if serr != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	return uid, nil
}

func (backend *SqliteBackend) setUserState(state int, nameuid string) *Error {
	if nameuid == "" {
		return &Error{"EINVAL", "User name or uid must be given"}
	}
	result, err := backend.setUserStateStmt.Exec(state, nameuid, nameuid)
	if err != nil {
		return &Error{"EFAULT", err.Error()}
	}
	n, aerr := result.RowsAffected()
	if aerr != nil {
		return &Error{"EFAULT", err.Error()}
	}
	if n == 0 {
		return &Error{"ENOENT", "User name"}
	}
	return nil
}
