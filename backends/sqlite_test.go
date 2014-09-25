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

func TestEnableDisableUser(t *testing.T) {
	backend, dberr := NewSqliteBackend(":memory:")
	if dberr != nil {
		t.Fatal(dberr)
	}
	defer backend.Close()

	// it rejects invalid calls
	inval := backend.DisableUser("")
	if inval.Code != "EINVAL" {
		t.Fatal("should return EINVAL instead of", inval.Code)
	}

	// fails to disable unknown user
	enoent := backend.DisableUser("test@example.com")
	if enoent.Code != "ENOENT" {
		t.Fatal("should fail to disable unkown user with code ENOENT but was",
			enoent.Code)
	}

	// create a user and check he can't login after he was disabled
	backend.CreateUser("test@example.com", "secret")
	_ = backend.DisableUser("test@example.com")
	_, err := backend.LoginUser("test@example.com", "secret")
	if err == nil {
		t.Fatal("should fail to login a user")
	}

	backend.EnableUser("test@example.com")
	uid, err := backend.LoginUser("test@example.com", "secret")
	if err != nil {
		t.Fatal("should fail to login a user")
	}
	if uid != 1 {
		t.Fatal("should have logged in the user after activation")
	}
}

func TestSetGetUserData(t *testing.T) {
	backend, dberr := NewSqliteBackend(":memory:")
	if dberr != nil {
		t.Fatal(dberr)
	}
	defer backend.Close()

	err := backend.SetUserData("test@example.com", "firstname", "Tester")
	if err.Code != "ENOENT" {
		t.Fatal("should fail to set value on non existing user", err.Code)
	}

	err1 := backend.SetUserData("1", "firstname", "Tester")
	if err1.Code != "ENOENT" {
		t.Fatal("should fail to set value on non existing user", err1.Code)
	}

	_, err2 := backend.GetUserData("test@example.com", "firstname")
	if err2.Code != "ENOENT" {
		t.Fatal("should fail to set value on non existing user", err2.Code)
	}

	_, err3 := backend.GetUserData("1", "firstname")
	if err3.Code != "ENOENT" {
		t.Fatal("should fail to set value on non existing user", err3.Code)
	}

	intval := backend.SetUserData("", "firstname", "Tester")
	if intval.Code != "EINVAL" {
		t.Fatal("should fail to set value on non invalid email", intval.Code)
	}

	intval1 := backend.SetUserData("test@example.com", "", "Tester")
	if intval1.Code != "EINVAL" {
		t.Fatal("should fail to set value on non invalid key", intval1.Code)
	}

	intval2 := backend.SetUserData("test@example.com", "firstname", "")
	if intval2.Code != "EINVAL" {
		t.Fatal("should fail to set value on non invalid value", intval2.Code)
	}

	_, intval3 := backend.GetUserData("", "firstname")
	if intval3.Code != "EINVAL" {
		t.Fatal("should fail to set value on non invalid value", intval3.Code)
	}

	_, intval4 := backend.GetUserData("test@example.com", "")
	if intval4.Code != "EINVAL" {
		t.Fatal("should fail to set value on non invalid value", intval4.Code)
	}

	backend.CreateUser("test@example.com", "secret")
	backend.SetUserData("test@example.com", "firstname", "Tester")
	val, _ := backend.GetUserData("test@example.com", "firstname")
	if val != "Tester" {
		t.Fatal("the value should have been 'Tester' but was", val)
	}
}

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
