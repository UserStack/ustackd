package connection

import (
	"fmt"
	"strings"

	"github.com/UserStack/ustackd/backends"
)

type Interpreter struct {
	*Context
}

func (ip *Interpreter) parse(line string) {
	ip.Log(line)
	if strings.HasPrefix(line, "client auth ") {
		ip.clientAuth(line[12:])
	} else if strings.HasPrefix(line, "login ") {
		ip.login(line[6:])
	} else if strings.HasPrefix(line, "disable ") {
		ip.disable(line[8:])
	} else if strings.HasPrefix(line, "enable ") {
		ip.enable(line[7:])
	} else if strings.HasPrefix(line, "set ") {
		ip.set(line[4:])
	} else if strings.HasPrefix(line, "get ") {
		ip.get(line[4:])
	} else if strings.HasPrefix(line, "change password ") {
		ip.changePassword(line[:16])
	} else if strings.HasPrefix(line, "change email ") {
		ip.changeEmail(line[:13])
	} else if strings.HasPrefix(line, "user groups ") {
		ip.userGroups(line[:12])
	} else if strings.HasPrefix(line, "user ") {
		ip.user(line[:5])
	} else if strings.HasPrefix(line, "delete user ") {
		ip.deleteUser(line[:12])
	} else if strings.HasPrefix(line, "users") {
		ip.users(line[:5])
	} else if strings.HasPrefix(line, "add ") {
		ip.add(line[:4])
	} else if strings.HasPrefix(line, "remove ") {
		ip.remove(line[:7])
	} else if strings.HasPrefix(line, "delete group ") {
		ip.deleteGroup(line[:13])
	} else if strings.HasPrefix(line, "groups") {
		ip.groups(line[:6])
	} else if strings.HasPrefix(line, "group users ") {
		ip.groupUsers(line[:12])
	} else if strings.HasPrefix(line, "group ") {
		ip.group(line[:6])
	} else if line == "quit" {
		ip.quit(line)
	} else if line == "stats" {
		ip.Ok()
	} else {
		ip.Err("EFAULT")
	}
}

func (ip *Interpreter) clientAuth(passwd string) {
	ip.Log("Try login with '" + passwd + "'")

	if passwd == "secret" {
		ip.loggedin = true
		ip.OkValue("(user group)")
	} else {
		ip.Err("EPERM")
	}
}

// login <email> <password>
func (ip *Interpreter) login(line string) {
	ip.withArgs(line, 2, func(args []string) {
		uid, err := ip.Backend.LoginUser(args[0], args[1])
		ip.intResponder(uid, err)
	})
}

// disable <email|uid>
func (ip *Interpreter) disable(emailuid string) {
	ip.simpleResponder(ip.Backend.DisableUser(emailuid))
}

// enable <email|uid>
func (ip *Interpreter) enable(emailuid string) {
	ip.simpleResponder(ip.Backend.EnableUser(emailuid))
}

// set <email|uid> <key> <value>
func (ip *Interpreter) set(line string) {
	ip.withArgs(line, 3, func(args []string) {
		ip.simpleResponder(
			ip.Backend.SetUserData(args[0], args[1], args[2]))
	})
}

// get <email|uid> <key>
func (ip *Interpreter) get(line string) {
	ip.withArgs(line, 2, func(args []string) {
		val, err := ip.Backend.GetUserData(args[0], args[1])
		if err == nil {
			ip.Write(val)
		}
		ip.simpleResponder(err)
	})
}

// change password <email|uid> <password> <newpassword>
func (ip *Interpreter) changePassword(line string) {
	ip.withArgs(line, 3, func(args []string) {
		ip.simpleResponder(
			ip.Backend.ChangeUserPassword(args[0], args[1], args[2]))
	})
}

// change email <email|uid> <password> <newemail>
func (ip *Interpreter) changeEmail(line string) {
	ip.withArgs(line, 3, func(args []string) {
		ip.simpleResponder(
			ip.Backend.ChangeUserEmail(args[0], args[1], args[2]))
	})
}

// user groups <email|uid>
func (ip *Interpreter) userGroups(emailuid string) {
	items, err := ip.Backend.UserGroups(emailuid)
	ip.groupResponder(items, err)
}

// user <email> <password>
func (ip *Interpreter) user(line string) {
	ip.withArgs(line, 2, func(args []string) {
		uid, err := ip.Backend.CreateUser(args[0], args[1])
		ip.intResponder(uid, err)
	})
}

// delete user <email|uid>
func (ip *Interpreter) deleteUser(emailuid string) {
	ip.simpleResponder(ip.Backend.DeleteUser(emailuid))
}

// users
func (ip *Interpreter) users(line string) {
	items, err := ip.Backend.Users()
	ip.userResponder(items, err)
}

// add <email|uid> to <group|gid>
func (ip *Interpreter) add(line string) {
	ip.withArgs(line, 2, func(args []string) {
		ip.simpleResponder(ip.Backend.AddUserToGroup(args[0], args[1]))
	})
}

// remove <email|uid> from <group|gid>
func (ip *Interpreter) remove(line string) {
	ip.withArgs(line, 2, func(args []string) {
		ip.simpleResponder(ip.Backend.RemoveUserFromGroup(args[0], args[1]))
	})
}

// delete group <group|gid>
func (ip *Interpreter) deleteGroup(groupgid string) {
	ip.simpleResponder(ip.Backend.DeleteGroup(groupgid))
}

// groups
func (ip *Interpreter) groups(line string) {
	items, err := ip.Backend.Groups()
	ip.groupResponder(items, err)
}

// group users <group|gid>
func (ip *Interpreter) groupUsers(groupgid string) {
	items, err := ip.Backend.GroupUsers(groupgid)
	ip.userResponder(items, err)
}

// group <name>
func (ip *Interpreter) group(name string) {
	gid, err := ip.Backend.Group(name)
	ip.intResponder(gid, err)
}

func (ip *Interpreter) quit(line string) {
	ip.quitting = true
	ip.Write("+ BYE")
}

// Helpers

type withArgsFn func([]string)

func (ip *Interpreter) withArgs(line string, n int, fn withArgsFn) {
	args := strings.SplitN(line, " ", n)
	if len(args) == n {
		fn(args)
	} else {
		ip.Err("EFAULT")
	}
}

func (ip *Interpreter) simpleResponder(err *backends.Error) {
	if err != nil {
		ip.Err(err.Code)
	} else {
		ip.Ok()
	}
}

func (ip *Interpreter) intResponder(value int64, err *backends.Error) {
	if err != nil {
		ip.Err(err.Code)
	} else {
		ip.OkValue(fmt.Sprintf("%d", value))
	}
}

func (ip *Interpreter) groupResponder(items []backends.Group, err *backends.Error) {
	if err != nil {
		ip.Err(err.Code)
	} else {
		for _, item := range items {
			ip.Write(item.String())
		}
		ip.Ok()
	}
}

func (ip *Interpreter) userResponder(items []backends.User, err *backends.Error) {
	if err != nil {
		ip.Err(err.Code)
	} else {
		for _, item := range items {
			ip.Write(item.String())
		}
		ip.Ok()
	}
}
