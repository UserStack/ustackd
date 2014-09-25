package backends

import (
	"fmt"
)

type User struct {
	Uid  int64
	Name string
}

func (u User) String() string {
	return fmt.Sprintf("%s:%d", u.Name, u.Uid)
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
	CreateUser(name string, password string) (int64, *Error)
	DisableUser(nameuid string) *Error
	EnableUser(nameuid string) *Error
	SetUserData(nameuid string, key string, value string) *Error
	GetUserData(nameuid string, key string) (string, *Error)
	LoginUser(name string, password string) (int64, *Error)
	ChangeUserPassword(nameuid string, password string, newpassword string) *Error
	ChangeUserName(nameuid string, password string, newname string) *Error
	UserGroups(nameuid string) ([]Group, *Error)
	DeleteUser(nameuid string) *Error
	Users() ([]User, *Error)
	Group(name string) (int64, *Error)
	AddUserToGroup(nameuid string, groupgid string) *Error
	RemoveUserFromGroup(nameuid string, groupgid string) *Error
	DeleteGroup(groupgid string) *Error
	Groups() ([]Group, *Error)
	GroupUsers(groupgid string) ([]User, *Error)
	Close()
}
