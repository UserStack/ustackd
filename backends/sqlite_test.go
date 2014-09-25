package backends

import (
	"testing"
)

func TestCreateUser(t *testing.T) {
	backend, dberr := NewSqliteBackend(":memory:")
	if dberr != nil {
		t.Fatal(dberr)
	}
	defer backend.Close()

	// works first time
	uid, _ := backend.CreateUser("test@example.com", "secret")
	if uid <= 0 {
		t.Fatal("expect uid to be greater than 0")
	}

	// but fails second time with EEXIST
	_, berr2 := backend.CreateUser("test@example.com", "secret")
	if berr2.Code != "EEXIST" {
		t.Fatal("should return EEXIST instead of", berr2.Code)
	}

	// and fails with invalid data password ...
	_, berr3 := backend.CreateUser("test@example.com", "")
	if berr3.Code != "EINVAL" {
		t.Fatal("should return EINVAL instead of", berr3.Code)
	}

	// ... or email
	_, berr4 := backend.CreateUser("", "secret")
	if berr4.Code != "EINVAL" {
		t.Fatal("should return EINVAL instead of", berr4.Code)
	}
}

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

func TestLoginUser(t *testing.T) {
	backend, dberr := NewSqliteBackend(":memory:")
	if dberr != nil {
		t.Fatal(dberr)
	}
	defer backend.Close()

	_, err := backend.LoginUser("", "secret")
	if err.Code != "EINVAL" {
		t.Fatal("should have detected a blank username", err.Code)
	}

	_, err1 := backend.LoginUser("test0@example.com", "")
	if err1.Code != "EINVAL" {
		t.Fatal("should have detected a blank password", err1.Code)
	}

	_, err2 := backend.LoginUser("test0@example.com", "secret")
	if err2.Code != "ENOENT" {
		t.Fatal("should have reported, that user is unknown but was", err2.Code)
	}

	uid, _ := backend.CreateUser("test@example.com", "secret")
	uid2, _ := backend.LoginUser("test@example.com", "secret")
	if uid != uid2 {
		t.Fatal("should have found the same user with uid", uid,
			"but found", uid2)
	}
}

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

func TestDeleteUser(t *testing.T) {
	backend, dberr := NewSqliteBackend(":memory:")
	if dberr != nil {
		t.Fatal(dberr)
	}
	defer backend.Close()

	err := backend.DeleteUser("")
	if err.Code != "EINVAL" {
		t.Fatal("should error on missing parameter")
	}

	// create two users
	backend.CreateUser("test0@example.com", "secret")
	backend.CreateUser("test1@example.com", "secret")

	// now has two users
	users, _ := backend.Users()
	if len(users) != 2 {
		t.Fatal("user count should have been 2 but was", len(users))
	}

	// delete one using email
	backend.DeleteUser("test1@example.com")
	users, _ = backend.Users()
	if len(users) != 1 {
		t.Fatal("user count should have been 1 but was", len(users))
	}

	// delete one using uid
	backend.DeleteUser("1")
	users, _ = backend.Users()
	if len(users) != 0 {
		t.Fatal("user count should have been 0 but was", len(users))
	}
}

// func (backend *SqliteBackend) DeleteUser(emailuid string) *Error {
//     return nil
// }

func TestUsers(t *testing.T) {
	backend, dberr := NewSqliteBackend(":memory:")
	if dberr != nil {
		t.Fatal(dberr)
	}
	defer backend.Close()

	users, _ := backend.Users()
	if len(users) != 0 {
		t.Fatal("users should be empty and is ", len(users))
	}

	// create a user
	uid, _ := backend.CreateUser("test@example.com", "secret")
	if uid <= 0 {
		t.Fatal("expect uid to be greater than 0")
	}

	// the list should have one enties now
	users, _ = backend.Users()
	if len(users) != 1 {
		t.Fatal("users should be empty and is ", len(users))
	}
	if users[0].Email != "test@example.com" {
		t.Fatal("email should have been 'test@example.com' but was", users[0].Email)
	}
	if users[0].Uid != 1 {
		t.Fatal("email should have been 1 but was", users[0].Uid)
	}
}

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
