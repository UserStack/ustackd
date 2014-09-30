package backends

import (
	"database/sql"
	"fmt"
)

const (
	STATUS_ACTIVE   = 1
	STATUS_INACTIVE = 0
)

type SqlBackend struct {
	db                      *sql.DB
	createUserStmt          *sql.Stmt
	usersStmt               *sql.Stmt
	deleteUserStmt          *sql.Stmt
	loginUserStmt           *sql.Stmt
	setUserStateStmt        *sql.Stmt
	uidForNameUidStmt       *sql.Stmt
	setUserDataStmt         *sql.Stmt
	getUserDataStmt         *sql.Stmt
	changeUserPasswordStmt  *sql.Stmt
	changeUserNameStmt      *sql.Stmt
	userGroupsStmt          *sql.Stmt
	createGroupStmt         *sql.Stmt
	groupsStmt              *sql.Stmt
	deleteGroupStmt         *sql.Stmt
	gidForNameGidStmt       *sql.Stmt
	addUserToGroupStmt      *sql.Stmt
	removeUserFromGroupStmt *sql.Stmt
	groupUsersStmt          *sql.Stmt
}

func (backend *SqlBackend) init(prepare []string) error {
	var err error
	// set the default encoding, enable foreign keys, enable journal mode,
	// check the integrity, set timeout to 60 sec and enable the auto vacuum
	// and initialize all tables
	for _, stmt := range prepare {
		_, err = backend.db.Exec(stmt)
		if err != nil {
			return err
		}
	}
	backend.createUserStmt, err = backend.db.Prepare(`INSERT INTO Users
		(name, password) VALUES (?, ?);`)
	if err != nil {
		return err
	}
	backend.usersStmt, err = backend.db.Prepare(`SELECT name, uid, state FROM Users`)
	if err != nil {
		return err
	}
	backend.deleteUserStmt, err = backend.db.Prepare(`DELETE FROM Users WHERE uid = ? OR name = ?;`)
	if err != nil {
		return err
	}
	backend.loginUserStmt, err = backend.db.Prepare(fmt.Sprintf(
		"SELECT uid FROM Users WHERE name = ? AND password = ? AND state = %d;",
		STATUS_ACTIVE))
	if err != nil {
		return err
	}
	backend.setUserStateStmt, err = backend.db.Prepare(`UPDATE Users
		SET state = ? WHERE name = ? OR uid = ?;`)
	if err != nil {
		return err
	}
	backend.uidForNameUidStmt, err = backend.db.Prepare(`SELECT uid FROM Users
		WHERE name = ? OR uid = ?;`)
	if err != nil {
		return err
	}
	backend.setUserDataStmt, err = backend.db.Prepare(`INSERT INTO UserValues
		(uid, key, value) VALUES (?, ?, ?);`)
	if err != nil {
		return err
	}
	backend.getUserDataStmt, err = backend.db.Prepare(`SELECT value FROM UserValues
		WHERE uid = ? AND key = ?;`)
	if err != nil {
		return err
	}
	backend.changeUserPasswordStmt, err = backend.db.Prepare(`UPDATE Users SET password = ?
		WHERE uid = ? AND password = ?;`)
	if err != nil {
		return err
	}
	backend.changeUserNameStmt, err = backend.db.Prepare(`UPDATE Users SET name = ?
		WHERE uid = ? AND password = ?;`)
	if err != nil {
		return err
	}
	backend.userGroupsStmt, err = backend.db.Prepare(`SELECT g.name, g.gid FROM Groups g
		JOIN UserGroups ug ON (ug.gid = g.gid)
		WHERE ug.uid = ?`)
	if err != nil {
		return err
	}
	backend.createGroupStmt, err = backend.db.Prepare(`INSERT INTO Groups (name)
		VALUES (?);`)
	if err != nil {
		return err
	}
	backend.groupsStmt, err = backend.db.Prepare(`SELECT name, gid FROM Groups;`)
	if err != nil {
		return err
	}
	backend.deleteGroupStmt, err = backend.db.Prepare(`DELETE FROM Groups
		WHERE gid = ? OR name = ?;`)
	if err != nil {
		return err
	}
	backend.gidForNameGidStmt, err = backend.db.Prepare(`SELECT gid FROM Groups
		WHERE name = ? OR gid = ?;`)
	if err != nil {
		return err
	}
	backend.addUserToGroupStmt, err = backend.db.Prepare(`INSERT INTO UserGroups (uid, gid)
		VALUES (?, ?);`)
	if err != nil {
		return err
	}
	backend.removeUserFromGroupStmt, err = backend.db.Prepare(`DELETE FROM UserGroups
		WHERE uid = ? AND gid = ?`)
	if err != nil {
		return err
	}
	backend.groupUsersStmt, err = backend.db.Prepare(`SELECT u.name, u.uid, u.state FROM Users u
		JOIN UserGroups ug ON (ug.uid = u.uid)
		WHERE ug.gid = ?`)
	if err != nil {
		return err
	}
	return nil
}

func (backend *SqlBackend) CreateUser(name string, password string) (int64, *Error) {
	if name == "" || password == "" {
		return 0, &Error{"EINVAL", "User name and password can't be blank"}
	}
	result, err := backend.createUserStmt.Exec(name, password)
	if err != nil {
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

func (backend *SqlBackend) DisableUser(nameuid string) *Error {
	return backend.setUserState(STATUS_INACTIVE, nameuid)
}

func (backend *SqlBackend) EnableUser(nameuid string) *Error {
	return backend.setUserState(STATUS_ACTIVE, nameuid)
}

func (backend *SqlBackend) SetUserData(nameuid string, key string, value string) *Error {
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

func (backend *SqlBackend) GetUserData(nameuid string, key string) (string, *Error) {
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

func (backend *SqlBackend) LoginUser(name string, password string) (int64, *Error) {
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

func (backend *SqlBackend) ChangeUserPassword(nameuid string, password string, newpassword string) *Error {
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

func (backend *SqlBackend) ChangeUserName(nameuid string, password string, newname string) *Error {
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

func (backend *SqlBackend) UserGroups(nameuid string) ([]Group, *Error) {
	if nameuid == "" {
		return nil, &Error{"EINVAL", "Name or uid has to be passed"}
	}
	uid, uerr := backend.getUidForNameUid(nameuid)
	if uerr != nil {
		return nil, uerr
	}
	var groups []Group
	rows, err := backend.userGroupsStmt.Query(uid)
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

func (backend *SqlBackend) DeleteUser(nameuid string) *Error {
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

func (backend *SqlBackend) Users() ([]User, *Error) {
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

func (backend *SqlBackend) CreateGroup(name string) (int64, *Error) {
	if name == "" {
		return 0, &Error{"EINVAL", "Invalid group name"}
	}
	result, err := backend.createGroupStmt.Exec(name)
	if err != nil {
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

func (backend *SqlBackend) AddUserToGroup(nameuid string, groupgid string) *Error {
	if nameuid == "" || groupgid == "" {
		return &Error{"EINVAL", "nameuid and groupgid can't be blank"}
	}
	uid, uerr := backend.getUidForNameUid(nameuid)
	if uerr != nil {
		return uerr
	}
	gid, gerr := backend.getGidForNameGid(groupgid)
	if gerr != nil {
		return gerr
	}
	_, aerr := backend.addUserToGroupStmt.Exec(uid, gid)
	if aerr != nil {
		return &Error{"EFAULT", aerr.Error()}
	}
	return nil
}

func (backend *SqlBackend) RemoveUserFromGroup(nameuid string, groupgid string) *Error {
	if nameuid == "" || groupgid == "" {
		return &Error{"EINVAL", "nameuid and groupgid can't be blank"}
	}
	uid, uerr := backend.getUidForNameUid(nameuid)
	if uerr != nil {
		return uerr
	}
	gid, gerr := backend.getGidForNameGid(groupgid)
	if gerr != nil {
		return gerr
	}
	_, aerr := backend.removeUserFromGroupStmt.Exec(uid, gid)
	if aerr != nil {
		return &Error{"EFAULT", aerr.Error()}
	}
	return nil
}

func (backend *SqlBackend) DeleteGroup(groupgid string) *Error {
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

func (backend *SqlBackend) Groups() ([]Group, *Error) {
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

func (backend *SqlBackend) GroupUsers(groupgid string) ([]User, *Error) {
	if groupgid == "" {
		return nil, &Error{"EINVAL", "Name or gid has to be passed"}
	}
	gid, gerr := backend.getGidForNameGid(groupgid)
	if gerr != nil {
		return nil, gerr
	}
	var users []User
	rows, err := backend.groupUsersStmt.Query(gid)
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

func (backend *SqlBackend) Close() {
	backend.db.Close()
}

func (backend *SqlBackend) getUidForNameUid(nameuid string) (int64, *Error) {
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

func (backend *SqlBackend) getGidForNameGid(groupgid string) (int64, *Error) {
	rows, err := backend.gidForNameGidStmt.Query(groupgid, groupgid)
	defer rows.Close()
	if err != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	if !rows.Next() {
		return 0, &Error{"ENOENT", "Name unknown"}
	}
	var gid int64
	serr := rows.Scan(&gid)
	if serr != nil {
		return 0, &Error{"EFAULT", err.Error()}
	}
	return gid, nil
}

func (backend *SqlBackend) setUserState(state int, nameuid string) *Error {
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
