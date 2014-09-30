package server

import (
	"regexp"
	"strings"

	"github.com/UserStack/ustackd/backends"
)

type Interpreter struct {
	*Context
	auth   *Auth
	regexp *regexp.Regexp
}

func (ip *Interpreter) parse(line string) {
	if !(ip.alwaysAvailableCommands(line) || ip.sensitiveCommands(line)) {
		ip.Err("EFAULT") // command unknown
	}
}

func (ip *Interpreter) alwaysAvailableCommands(line string) bool {
	cmd := strings.ToLower(line)
	if strings.HasPrefix(cmd, "client auth ") {
		ip.clientAuth(line[12:])
	} else if cmd == "quit" {
		ip.quit(line)
	} else {
		return false
	}
	return true
}

func (ip *Interpreter) sensitiveCommands(line string) bool {
	if !ip.authorized(line) {
		ip.Err("EACCES")
		return true
	}
	cmd := strings.ToLower(line)
	if strings.HasPrefix(cmd, "login ") {
		ip.login(line[6:])
	} else if strings.HasPrefix(cmd, "disable ") {
		ip.disable(line[8:])
	} else if strings.HasPrefix(cmd, "enable ") {
		ip.enable(line[7:])
	} else if strings.HasPrefix(cmd, "set ") {
		ip.set(line[4:])
	} else if strings.HasPrefix(cmd, "get ") {
		ip.get(line[4:])
	} else if strings.HasPrefix(cmd, "change password ") {
		ip.changePassword(line[16:])
	} else if strings.HasPrefix(cmd, "change name ") {
		ip.changeName(line[12:])
	} else if strings.HasPrefix(cmd, "user groups ") {
		ip.userGroups(line[12:])
	} else if strings.HasPrefix(cmd, "user ") {
		ip.user(line[5:])
	} else if strings.HasPrefix(cmd, "delete user ") {
		ip.deleteUser(line[12:])
	} else if strings.HasPrefix(cmd, "users") {
		ip.users(line[5:])
	} else if strings.HasPrefix(cmd, "add ") {
		ip.add(line[4:])
	} else if strings.HasPrefix(cmd, "remove ") {
		ip.remove(line[7:])
	} else if strings.HasPrefix(cmd, "delete group ") {
		ip.deleteGroup(line[13:])
	} else if strings.HasPrefix(cmd, "groups") {
		ip.groups(line[6:])
	} else if strings.HasPrefix(cmd, "group users ") {
		ip.groupUsers(line[12:])
	} else if strings.HasPrefix(cmd, "group ") {
		ip.group(line[6:])
	} else if cmd == "stats" {
		ip.stats(line[5:])
	} else {
		return false
	}
	return true
}

func (ip *Interpreter) authorized(line string) bool {
	if (len(ip.Cfg.Client.Auth)) > 0 { // auth strings are defined, auth required
		if ip.auth != nil { // is logged in
			if ip.auth.Allow && ip.regexp.MatchString(line) {
				return true
			} else if !ip.auth.Allow && !ip.regexp.MatchString(line) { // Deny
				return true
			} else {
				return false
			}
		} else {
			return false
		}
	} else {
		return true // no auth defined everthing is allowed
	}
}

func (ip *Interpreter) clientAuth(passwd string) {
	for _, auth := range ip.Cfg.Client.Auth {
		if auth.Passwd == passwd {
			ip.auth = &auth
			var err error
			ip.regexp, err = regexp.Compile(auth.Regex)
			if err != nil {
				ip.Log(err.Error())
				ip.Err("EFAULT")
				return
			}
			ip.Ok()
			return
		}
	}
	ip.Err("EPERM")
}

func (ip *Interpreter) stats(line string) {
	ip.Writef("Successfull logins: %d", ip.Server.Stats.Login)
	ip.Writef("Failed logins: %d", ip.Server.Stats.FailedLogin)
	ip.Ok()
}

// login <name> <password>
func (ip *Interpreter) login(line string) {
	ip.withArgs(line, 2, func(args []string) {
		uid, err := ip.Backend.LoginUser(args[0], args[1])
		if err == nil {
			ip.Server.Stats.Login++
		} else {
			ip.Server.Stats.FailedLogin++
		}
		ip.intResponder(uid, err)
	})
}

// disable <name|uid>
func (ip *Interpreter) disable(nameuid string) {
	ip.simpleResponder(ip.Backend.DisableUser(nameuid))
}

// enable <name|uid>
func (ip *Interpreter) enable(nameuid string) {
	ip.simpleResponder(ip.Backend.EnableUser(nameuid))
}

// set <name|uid> <key> <value>
func (ip *Interpreter) set(line string) {
	ip.withArgs(line, 3, func(args []string) {
		ip.simpleResponder(
			ip.Backend.SetUserData(args[0], args[1], args[2]))
	})
}

// get <name|uid> <key>
func (ip *Interpreter) get(line string) {
	ip.withArgs(line, 2, func(args []string) {
		val, err := ip.Backend.GetUserData(args[0], args[1])
		if err == nil {
			ip.Write(val)
		}
		ip.simpleResponder(err)
	})
}

// change password <name|uid> <password> <newpassword>
func (ip *Interpreter) changePassword(line string) {
	ip.withArgs(line, 3, func(args []string) {
		ip.simpleResponder(
			ip.Backend.ChangeUserPassword(args[0], args[1], args[2]))
	})
}

// change name <name|uid> <password> <newname>
func (ip *Interpreter) changeName(line string) {
	ip.withArgs(line, 3, func(args []string) {
		ip.simpleResponder(
			ip.Backend.ChangeUserName(args[0], args[1], args[2]))
	})
}

// user groups <name|uid>
func (ip *Interpreter) userGroups(nameuid string) {
	items, err := ip.Backend.UserGroups(nameuid)
	ip.groupResponder(items, err)
}

// user <name> <password>
func (ip *Interpreter) user(line string) {
	ip.withArgs(line, 2, func(args []string) {
		uid, err := ip.Backend.CreateUser(args[0], args[1])
		ip.intResponder(uid, err)
	})
}

// delete user <name|uid>
func (ip *Interpreter) deleteUser(nameuid string) {
	ip.simpleResponder(ip.Backend.DeleteUser(nameuid))
}

// users
func (ip *Interpreter) users(line string) {
	items, err := ip.Backend.Users()
	ip.userResponder(items, err)
}

// add <name|uid> to <group|gid>
func (ip *Interpreter) add(line string) {
	ip.withArgs(line, 2, func(args []string) {
		ip.simpleResponder(ip.Backend.AddUserToGroup(args[0], args[1]))
	})
}

// remove <name|uid> from <group|gid>
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
	gid, err := ip.Backend.CreateGroup(name)
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
		ip.Err("EINVAL")
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
		ip.OkValue(value)
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
