package connection

import (
	"strconv"
	"strings"

	"github.com/UserStack/ustackd/backends"
)

type Interpreter struct {
	context *Context
	backend backends.Abstract
}

func (interpreter *Interpreter) parse(line string) {
	interpreter.context.Log(line)
	if strings.HasPrefix(line, "client auth ") {
		interpreter.clientAuth(line[12:])
	} else if strings.HasPrefix(line, "login ") {
		interpreter.login(line[6:])
	} else if strings.HasPrefix(line, "disable ") {
		interpreter.disable(line[8:])
	} else if strings.HasPrefix(line, "enable ") {
		interpreter.enable(line[7:])
	} else if strings.HasPrefix(line, "set ") {
		interpreter.set(line[4:])
	} else if strings.HasPrefix(line, "get ") {
		interpreter.get(line[4:])
	} else if strings.HasPrefix(line, "change password ") {
		interpreter.changePassword(line[:16])
	} else if strings.HasPrefix(line, "change email ") {
		interpreter.changeEmail(line[:13])
	} else if strings.HasPrefix(line, "user groups ") {
		interpreter.userGroups(line[:12])
	} else if strings.HasPrefix(line, "user ") {
		interpreter.user(line[:5])
	} else if strings.HasPrefix(line, "delete user ") {
		interpreter.deleteUser(line[:12])
	} else if strings.HasPrefix(line, "users") {
		interpreter.users(line[:5])
	} else if strings.HasPrefix(line, "add ") {
		interpreter.add(line[:4])
	} else if strings.HasPrefix(line, "remove ") {
		interpreter.remove(line[:7])
	} else if strings.HasPrefix(line, "delete group ") {
		interpreter.deleteGroup(line[:13])
	} else if strings.HasPrefix(line, "groups") {
		interpreter.groups(line[:6])
	} else if strings.HasPrefix(line, "group users ") {
		interpreter.groupUsers(line[:12])
	} else if strings.HasPrefix(line, "group ") {
		interpreter.group(line[:6])
	} else if line == "quit" {
		interpreter.quit(line)
	} else if line == "stats" {
		interpreter.context.Ok()
	} else {
		interpreter.context.Err("EFAULT")
	}
}

func (interpreter *Interpreter) clientAuth(passwd string) {
	interpreter.context.Log("Try login with '" + passwd + "'")

	if passwd == "secret" {
		interpreter.context.loggedin = true
		interpreter.context.OkValue("(user group)")
	} else {
		interpreter.context.Err("EPERM")
	}
}

// login <email> <password>
func (interpreter *Interpreter) login(line string) {
	interpreter.withArgs(line, 2, func(args []string) {
		uid, err := interpreter.backend.LoginUser(args[0], args[1])
		interpreter.intResponder(uid, err)
	})
}

// disable <email|uid>
func (interpreter *Interpreter) disable(emailuid string) {
	interpreter.simpleResponder(interpreter.backend.DisableUser(emailuid))
}

// enable <email|uid>
func (interpreter *Interpreter) enable(emailuid string) {
	interpreter.simpleResponder(interpreter.backend.EnableUser(emailuid))
}

// set <email|uid> <key> <value>
func (interpreter *Interpreter) set(line string) {
	interpreter.withArgs(line, 3, func(args []string) {
		interpreter.simpleResponder(
			interpreter.backend.SetUserData(args[0], args[1], args[2]))
	})
}

// get <email|uid> <key>
func (interpreter *Interpreter) get(line string) {
	interpreter.withArgs(line, 2, func(args []string) {
		interpreter.simpleResponder(interpreter.backend.GetUserData(args[0], args[1]))
	})
}

// change password <email|uid> <password> <newpassword>
func (interpreter *Interpreter) changePassword(line string) {
	interpreter.withArgs(line, 3, func(args []string) {
		interpreter.simpleResponder(
			interpreter.backend.ChangeUserPassword(args[0], args[1], args[2]))
	})
}

// change email <email|uid> <password> <newemail>
func (interpreter *Interpreter) changeEmail(line string) {
	interpreter.withArgs(line, 3, func(args []string) {
		interpreter.simpleResponder(
			interpreter.backend.ChangeUserEmail(args[0], args[1], args[2]))
	})
}

// user groups <email|uid>
func (interpreter *Interpreter) userGroups(emailuid string) {
	items, err := interpreter.backend.UserGroups(emailuid)

	if err != nil {
		interpreter.context.Err(err.Code)
	} else {
		for _, item := range items {
			interpreter.context.Write(item.Line())
		}
		interpreter.context.Ok()
	}
}

// user <email> <password>
func (interpreter *Interpreter) user(line string) {
	interpreter.withArgs(line, 2, func(args []string) {
		uid, err := interpreter.backend.CreateUser(args[0], args[1])
		interpreter.intResponder(uid, err)
	})
}

// delete user <email|uid>
func (interpreter *Interpreter) deleteUser(emailuid string) {
	interpreter.simpleResponder(interpreter.backend.DeleteUser(emailuid))
}

// users
func (interpreter *Interpreter) users(line string) {
	items, err := interpreter.backend.Users()

	if err != nil {
		interpreter.context.Err(err.Code)
	} else {
		for _, item := range items {
			interpreter.context.Write(item.Line())
		}
		interpreter.context.Ok()
	}
}

// add <email|uid> to <group|gid>
func (interpreter *Interpreter) add(line string) {
	interpreter.withArgs(line, 2, func(args []string) {
		interpreter.simpleResponder(interpreter.backend.AddUserToGroup(args[0], args[1]))
	})
}

// remove <email|uid> from <group|gid>
func (interpreter *Interpreter) remove(line string) {
	interpreter.withArgs(line, 2, func(args []string) {
		interpreter.simpleResponder(interpreter.backend.RemoveUserFromGroup(args[0], args[1]))
	})
}

// delete group <group|gid>
func (interpreter *Interpreter) deleteGroup(groupgid string) {
	interpreter.simpleResponder(interpreter.backend.DeleteGroup(groupgid))
}

// groups
func (interpreter *Interpreter) groups(line string) {
	items, err := interpreter.backend.Groups()

	if err != nil {
		interpreter.context.Err(err.Code)
	} else {
		for _, item := range items {
			interpreter.context.Write(item.Line())
		}
		interpreter.context.Ok()
	}
}

// group users <group|gid>
func (interpreter *Interpreter) groupUsers(groupgid string) {
	items, err := interpreter.backend.GroupUsers(groupgid)

	if err != nil {
		interpreter.context.Err(err.Code)
	} else {
		for _, item := range items {
			interpreter.context.Write(item.Line())
		}
		interpreter.context.Ok()
	}
}

// group <name>
func (interpreter *Interpreter) group(name string) {
	gid, err := interpreter.backend.Group(name)
	interpreter.intResponder(gid, err)
}

func (interpreter *Interpreter) quit(line string) {
	interpreter.context.quitting = true
	interpreter.context.Write("+ BYE")
}

// Helpers

type withArgsFn func([]string)

func (interpreter *Interpreter) withArgs(line string, n int, fn withArgsFn) {
	args := strings.SplitN(line, " ", n)
	if len(args) == n {
		fn(args)
	} else {
		interpreter.context.Err("EFAULT")
	}
}

func (interpreter *Interpreter) simpleResponder(err *backends.Error) {
	if err != nil {
		interpreter.context.Err(err.Code)
	} else {
		interpreter.context.Ok()
	}
}

func (interpreter *Interpreter) intResponder(value int, err *backends.Error) {
	if err != nil {
		interpreter.context.Err(err.Code)
	} else {
		interpreter.context.OkValue(strconv.Itoa(value))
	}
}
