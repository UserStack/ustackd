package backends

import (
	"fmt"
)

type User struct {
	Uid   int64
	Email string
}

func (u User) String() string {
	return fmt.Sprintf("%s:%d", u.Email, u.Uid)
}

type Group struct {
	Gid  int64
	Name string
}

func (g Group) String() string {
	return fmt.Sprintf("%s:%d", g.Name, g.Gid)
}

type Error struct {
	Code    string
	Message string
}

type Abstract interface {
	CreateUser(email string, password string) (int64, *Error)
	DisableUser(emailuid string) *Error
	EnableUser(emailuid string) *Error
	SetUserData(emailuid string, key string, value string) *Error
	GetUserData(emailuid string, key string) (string, *Error)
	LoginUser(email string, password string) (int64, *Error)
	ChangeUserPassword(emailuid string, password string, newpassword string) *Error
	ChangeUserEmail(emailuid string, password string, newemail string) *Error
	UserGroups(emailuid string) ([]Group, *Error)
	DeleteUser(emailuid string) *Error
	Users() ([]User, *Error)
	Group(name string) (int64, *Error)
	AddUserToGroup(emailuid string, groupgid string) *Error
	RemoveUserFromGroup(emailuid string, groupgid string) *Error
	DeleteGroup(groupgid string) *Error
	Groups() ([]Group, *Error)
	GroupUsers(groupgid string) ([]User, *Error)
	Close()
}
