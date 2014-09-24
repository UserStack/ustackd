package backends

import (
	"testing"
)

func TestCreateUser(t *testing.T) {
	backend, err := NewSqliteBackend(":memory:")
	if err != nil {
		t.Error(err)
	}
	// defer backend.Close()

	// works first time
	uid, berr := backend.CreateUser("test@example.com", "secret")
	if uid <= 0 || berr != nil {
		t.Error(berr, "also expect uid to be greater than 0")
	}

	// but fails second time with EEXIST
	uid, berr2 := backend.CreateUser("test@example.com", "secret")
	if uid > 0 || berr2.Code != "EEXIST" {
		t.Error("Should return EEXIST instead of", berr2.Code)
	}

	// and fails with invalid data password ...
	uid, berr3 := backend.CreateUser("test@example.com", "")
	if uid > 0 || berr3.Code != "EINVAL" {
		t.Error("Should return EINVAL instead of", berr3.Code)
	}

	// ... or email
	uid, berr4 := backend.CreateUser("", "secret")
	if uid > 0 || berr4.Code != "EINVAL" {
		t.Error("Should return EINVAL instead of", berr4.Code)
	}
}

// func (backend *SqliteBackend) CreateUser(email string, password string) (int, *Error) {
//     return 0, nil
// }
//
// func (backend *SqliteBackend) DisableUser(emailuid string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) EnableUser(emailuid string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) SetUserData(emailuid string, key string, value string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) GetUserData(emailuid string, key string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) LoginUser(email string, password string) (int, *Error) {
//     return 0, nil
// }
//
// func (backend *SqliteBackend) ChangeUserPassword(emailuid string, password string, newpassword string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) ChangeUserEmail(emailuid string, password string, newemail string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) UserGroups(emailuid string) ([]Group, *Error) {
//     return nil, nil
// }
//
// func (backend *SqliteBackend) DeleteUser(emailuid string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) Users() ([]User, *Error) {
//     return nil, nil
// }
//
// func (backend *SqliteBackend) Group(name string) (int, *Error) {
//     return 0, nil
// }
//
// func (backend *SqliteBackend) AddUserToGroup(emailuid string, groupgid string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) RemoveUserFromGroup(emailuid string, groupgid string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) DeleteGroup(groupgid string) *Error {
//     return nil
// }
//
// func (backend *SqliteBackend) Groups() ([]Group, *Error) {
//     return nil, nil
// }
//
// func (backend *SqliteBackend) GroupUsers(groupgid string) ([]User, *Error) {
//     return nil, nil
// }
