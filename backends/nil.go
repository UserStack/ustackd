package Backends

type NilBackend struct {
}

func (backend *NilBackend) CreateUser(email string, password string) (int, *Error) {
	return 0, nil
}

func (backend *NilBackend) DisableUser(emailuid string) *Error {
	return nil
}

func (backend *NilBackend) EnableUser(emailuid string) *Error {
	return nil
}

func (backend *NilBackend) SetUserData(emailuid string, key string, value string) *Error {
	return nil
}

func (backend *NilBackend) GetUserData(emailuid string, key string) *Error {
	return nil
}

func (backend *NilBackend) LoginUser(email string, password string) (int, *Error) {
	return 0, nil
}

func (backend *NilBackend) ChangeUserPassword(emailuid string, password string, newpassword string) *Error {
	return nil
}

func (backend *NilBackend) ChangeUserEmail(emailuid string, password string, newemail string) *Error {
	return nil
}

func (backend *NilBackend) UserGroups(email string, uid string) ([]Group, *Error) {
	return nil, nil
}

func (backend *NilBackend) DeleteUser(email string, uid string) *Error {
	return nil
}

func (backend *NilBackend) Users() ([]User, *Error) {
	return nil, nil
}

func (backend *NilBackend) Group(name string) (int, *Error) {
	return 0, nil
}

func (backend *NilBackend) AddUserToGroup(emailuid string, groupgid string) *Error {
	return nil
}

func (backend *NilBackend) RemoveUserFromGroup(emailuid string, groupgid string) *Error {
	return nil
}

func (backend *NilBackend) DeleteGroup(groupgid string) *Error {
	return nil
}

func (backend *NilBackend) Groups() ([]Group, *Error) {
	return nil, nil
}

func (backend *NilBackend) GroupUsers(groupgid string) ([]User, *Error) {
	return nil, nil
}
