package connection

import (
	"strconv"
	"strings"
)

func clientAuth(context *Context, passwd string) {
	context.Log("Try login with '" + passwd + "'")

	if passwd == "secret" {
		context.Loggedin = true
		context.OkValue("(user group)")
	} else {
		context.Err("EPERM")
	}
}

// login <email> <password>
func login(context *Context, emailPassword string) {
	args := strings.SplitN(emailPassword, " ", 2)
	uid, err := context.Backend.LoginUser(args[0], args[1])

	if err != nil {
		context.Err(err.Code)
	} else {
		context.OkValue(strconv.Itoa(uid))
	}
}

// disable <email|uid>
func disable(context *Context, emailuid string) {
	context.Backend.DisableUser(emailuid)
}

func quit(context *Context, line string) {
	context.Quitting = true
	context.Write("+ BYE")
}

func interpret(context *Context, line string) {
	context.Log(line)
	if strings.HasPrefix(line, "client auth ") {
		clientAuth(context, line[12:])
	} else if strings.HasPrefix(line, "login ") {
		login(context, line[6:])
	} else if strings.HasPrefix(line, "disable ") {
		disable(context, line[8:])
	} else if strings.HasPrefix(line, "enable ") {
		// enable <email|uid>
		var emailuid string
		context.Backend.EnableUser(emailuid)
	} else if strings.HasPrefix(line, "set ") {
		// set <email|uid> <key> <value>
		var emailuid, key, value string
		context.Backend.SetUserData(emailuid, key, value)
	} else if strings.HasPrefix(line, "get ") {
		// get <email|uid> <key>
		var emailuid, key string
		context.Backend.GetUserData(emailuid, key)
	} else if strings.HasPrefix(line, "change password ") {
		// change password <email|uid> <password> <newpassword>
		var emailuid, password, newpassword string
		context.Backend.ChangeUserPassword(emailuid, password, newpassword)
	} else if strings.HasPrefix(line, "change email ") {
		// change email <email|uid> <password> <newemail>
		var emailuid, password, newemail string
		context.Backend.ChangeUserEmail(emailuid, password, newemail)
	} else if strings.HasPrefix(line, "user groups ") {
		// user groups <email|uid>
		var emailuid string
		context.Backend.UserGroups(emailuid)
	} else if strings.HasPrefix(line, "user ") {
		// user <email> <password>
		var email, password string
		context.Backend.CreateUser(email, password)
	} else if strings.HasPrefix(line, "delete user ") {
		// delete user <email|uid>
		var emailuid string
		context.Backend.DeleteUser(emailuid)
	} else if strings.HasPrefix(line, "users") {
		// users
		context.Backend.Users()
	} else if strings.HasPrefix(line, "add ") {
		// add <email|uid> to <group|gid>
		var emailuid, groupgid string
		context.Backend.AddUserToGroup(emailuid, groupgid)
	} else if strings.HasPrefix(line, "remove ") {
		// remove <email|uid> from <group|gid>
		var emailuid, groupgid string
		context.Backend.RemoveUserFromGroup(emailuid, groupgid)
	} else if strings.HasPrefix(line, "delete group ") {
		// delete group <group|gid>
		var groupgid string
		context.Backend.DeleteGroup(groupgid)
	} else if strings.HasPrefix(line, "groups") {
		// groups
		context.Backend.Groups()
	} else if strings.HasPrefix(line, "group users ") {
		// group users <group|gid>
		var groupgid string
		context.Backend.GroupUsers(groupgid)
	} else if strings.HasPrefix(line, "group ") {
		// group <name>
		var name string
		context.Backend.Group(name)
	} else if line == "quit" {
		quit(context, line)
	} else if line == "stats" {
		context.Ok()
	} else {
		context.Err("EFAULT")
	}
}
