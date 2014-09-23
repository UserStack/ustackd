package Backends

type User struct {
	uid   int
	email string
}

type Group struct {
	gid  int
	name string
}

type Error struct {
	code string
	msg  string
}

type Abstract interface {
	CreateUser(email string, password string) (int, *Error)
	DisableUser(emailuid string) *Error
	EnableUser(emailuid string) *Error
	SetUserData(emailuid string, key string, value string) *Error
	GetUserData(emailuid string, key string) *Error
	LoginUser(email string, password string) (int, *Error)
	ChangeUserPassword(emailuid string, password string, newpassword string) *Error
	ChangeUserEmail(emailuid string, password string, newemail string) *Error
	UserGroups(email string, uid string) ([]Group, *Error)
	DeleteUser(email string, uid string) *Error
	Users() ([]User, *Error)
	Group(name string) (int, *Error)
	AddUserToGroup(emailuid string, groupgid string) *Error
	RemoveUserFromGroup(emailuid string, groupgid string) *Error
	DeleteGroup(groupgid string) *Error
	Groups() ([]Group, *Error)
	GroupUsers(groupgid string) ([]User, *Error)
}
