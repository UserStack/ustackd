package server

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/UserStack/ustackd/backends"
)

type Interpreter struct {
	*Context
	auth   *Auth
	regexp *regexp.Regexp
}

func (ip *Interpreter) parse(line string) {
	cmd, args := parseCmd(line)
	switch cmd {
	case ERR_UNKNOWN_FUNC:
		ip.Err("EFAULT")
	case ERR_MISSING_ARGS, ERR_INVALID_ARGS:
		ip.Err("EINVAL")
	case CLIENT_AUTH, QUIT:
		ip.unrestrictedCommands(cmd, args)
	default:
		if !ip.authorized(strings.ToLower(line)) {
			ip.Err("EACCES")
			ip.Server.Stats.restrictedCommandsAccessDenied++
			return
		}
		ip.restrictedCommands(cmd, args)
	}
}

func (ip *Interpreter) unrestrictedCommands(cmd Command, args []string) {
	ip.Server.Stats.unrestrictedCommands++
	switch cmd {
	case CLIENT_AUTH:
		ip.clientAuth(args)
	case QUIT:
		ip.quit()
	}
}

func (ip *Interpreter) restrictedCommands(cmd Command, args []string) {
	ip.Server.Stats.restrictedCommands++
	switch cmd {
	case LOGIN:
		ip.login(args)
	case DISABLE:
		ip.disable(args)
	case ENABLE:
		ip.enable(args)
	case SET:
		ip.set(args)
	case GET:
		ip.get(args)
	case GETKEYS:
		ip.getKeys(args)
	case CHANGE_PASSWORD:
		ip.changePassword(args)
	case CHANGE_NAME:
		ip.changeName(args)
	case USER_GROUPS:
		ip.userGroups(args)
	case USER:
		ip.user(args)
	case DELETE_USER:
		ip.deleteUser(args)
	case USERS:
		ip.users()
	case ADD:
		ip.add(args)
	case REMOVE:
		ip.remove(args)
	case DELETE_GROUP:
		ip.deleteGroup(args)
	case GROUPS:
		ip.groups()
	case GROUP_USERS:
		ip.groupUsers(args)
	case GROUP:
		ip.group(args)
	case STATS:
		ip.stats()
	case LOGINSTATS:
		ip.loginStats(args)
	}

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

func (ip *Interpreter) clientAuth(passwd []string) {
	for _, auth := range ip.Cfg.Client.Auth {
		if auth.Passwd == passwd[0] {
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

func (ip *Interpreter) stats() {
	ip.Writef("Connects: %d", ip.Server.Stats.Connects)
	ip.Writef("Disconnects: %d", ip.Server.Stats.Disconnects)
	ip.Writef("Active Connections: %d", ip.Server.Stats.ActiveConnections())
	ip.Writef("Successfull logins: %d", ip.Server.Stats.Login)
	ip.Writef("Failed logins: %d", ip.Server.Stats.FailedLogin)
	ip.Writef("Unrestricted Commands: %d", ip.Server.Stats.unrestrictedCommands)
	ip.Writef("Restricted Commands: %d", ip.Server.Stats.restrictedCommands)
	ip.Writef("Access denied on Restricted Commands: %d", ip.Server.Stats.restrictedCommandsAccessDenied)

	stats, err := ip.Backend.Stats()
	if err != nil {
		ip.Err(err.Code)
		return
	}
	for key, value := range stats {
		ip.Writef("%s: %d", key, value)
	}
	ip.Ok()
}

func (ip *Interpreter) loginStats(args []string) {
	lastS, err := ip.Backend.GetUserData(args[0], "lastlogin")
	if err != nil {
		ip.Err(err.Code)
		return
	}
	last, serr := strconv.ParseInt(lastS, 10, 0)
	if serr != nil {
		ip.Err("NOINT")
		return
	}

	ip.Writef("Last successfull login: %s", time.Unix(last, 0))

	countS, err := ip.Backend.GetUserData(args[0], "failcount")
	if err != nil {
		ip.Err(err.Code)
		return
	}
	ip.Writef("Failed login attempts: %s", countS)
	ip.Ok()
}

// login <name> <password>
func (ip *Interpreter) login(args []string) {
	uid, err := ip.Backend.LoginUser(args[0], args[1])
	if err == nil {
		ip.Server.Stats.Login++
	} else {
		ip.Server.Stats.FailedLogin++
	}
	ip.intResponder(uid, err)
}

// disable <name|uid>
func (ip *Interpreter) disable(args []string) {
	ip.simpleResponder(ip.Backend.DisableUser(args[0]))
}

// enable <name|uid>
func (ip *Interpreter) enable(args []string) {
	ip.simpleResponder(ip.Backend.EnableUser(args[0]))
}

// set <name|uid> <key> <value>
func (ip *Interpreter) set(args []string) {
	ip.simpleResponder(ip.Backend.SetUserData(args[0], args[1], args[2]))
}

// get <name|uid> <key>
func (ip *Interpreter) get(args []string) {
	val, err := ip.Backend.GetUserData(args[0], args[1])
	if err == nil {
		ip.Write(val)
	}
	ip.simpleResponder(err)
}

// getkeys <name|uid>
func (ip *Interpreter) getKeys(args []string) {
	list, err := ip.Backend.GetUserDataKeys(args[0])
	if err == nil {
		for _, key := range list {
			ip.Write(key)
		}
	}
	ip.simpleResponder(err)
}

// change password <name|uid> <password> <newpassword>
func (ip *Interpreter) changePassword(args []string) {
	ip.simpleResponder(ip.Backend.ChangeUserPassword(args[0], args[1], args[2]))
}

// change name <name|uid> <password> <newname>
func (ip *Interpreter) changeName(args []string) {
	ip.simpleResponder(ip.Backend.ChangeUserName(args[0], args[1], args[2]))
}

// user groups <name|uid>
func (ip *Interpreter) userGroups(args []string) {
	items, err := ip.Backend.UserGroups(args[0])
	ip.groupResponder(items, err)
}

// user <name> <password>
func (ip *Interpreter) user(args []string) {
	uid, err := ip.Backend.CreateUser(args[0], args[1])
	ip.intResponder(uid, err)
}

// delete user <name|uid>
func (ip *Interpreter) deleteUser(args []string) {
	ip.simpleResponder(ip.Backend.DeleteUser(args[0]))
}

// users
func (ip *Interpreter) users() {
	items, err := ip.Backend.Users()
	ip.userResponder(items, err)
}

// add <name|uid> to <group|gid>
func (ip *Interpreter) add(args []string) {
	ip.simpleResponder(ip.Backend.AddUserToGroup(args[0], args[1]))
}

// remove <name|uid> from <group|gid>
func (ip *Interpreter) remove(args []string) {
	ip.simpleResponder(ip.Backend.RemoveUserFromGroup(args[0], args[1]))
}

// delete group <group|gid>
func (ip *Interpreter) deleteGroup(args []string) {
	ip.simpleResponder(ip.Backend.DeleteGroup(args[0]))
}

// groups
func (ip *Interpreter) groups() {
	items, err := ip.Backend.Groups()
	ip.groupResponder(items, err)
}

// group users <group|gid>
func (ip *Interpreter) groupUsers(args []string) {
	items, err := ip.Backend.GroupUsers(args[0])
	ip.userResponder(items, err)
}

// group <name>
func (ip *Interpreter) group(args []string) {
	gid, err := ip.Backend.CreateGroup(args[0])
	ip.intResponder(gid, err)
}

func (ip *Interpreter) quit() {
	ip.quitting = true
	ip.Write("+ BYE")
}

// Helpers

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
