package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	// "github.com/UserStack/ustackd/config"
	"github.com/UserStack/ustackd/backends"
)

// Client

type ConnectionContext struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	backend  *Backends.Abstract
	loggedin bool
	quitting bool
}

func (context *ConnectionContext) Write(line string) {
	context.writer.WriteString(line + "\r\n")
	context.writer.Flush()
}

func (context *ConnectionContext) Ok() {
	context.Write("+ OK")
}

func (context *ConnectionContext) OkValue(value string) {
	context.Write("+ OK " + value)
}

func (context *ConnectionContext) Err(code string) {
	context.Write("- " + code)
}

func (context *ConnectionContext) Log(line string) {
	fmt.Printf("%s: %s\r\n", context.conn.RemoteAddr(), line)
}

func (context *ConnectionContext) Realm() {
	realm := "ustackd 0.0.1"
	context.Log("new client connected")
	context.Write(realm + " (user group)")
}

func (context *ConnectionContext) Close() {
	context.conn.Close()
	context.Log("Client disonnected")
}

func NewConnectionContext(conn net.Conn, backend *Backends.Abstract) *ConnectionContext {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &ConnectionContext{conn, reader, writer, backend, false, false}
}

func clientAuth(context *ConnectionContext, passwd string) {
	context.Log("Try login with '" + passwd + "'")

	if passwd == "secret" {
		context.loggedin = true
		context.OkValue("(user group)")
	} else {
		context.Err("EPERM")
	}
}

func quit(context *ConnectionContext, line string) {
	context.quitting = true
	context.Write("+ BYE")
}

func interpret(context *ConnectionContext, line string) {
	context.Log(line)

	/*
		TODO:
		user <email> <password>			-> Backend.createUser(email, password)
		disable <email|uid> 				-> Backend.disableUser(emailuid)
		enable <email|uid> 				-> Backend.enableUser(emailuid)
		set <email|uid> <key> <value> 		-> Backend.setUserData(emailuid, key, value)
		get <email|uid> <key> 				-> Backend.getUserData(emailuid, key)
		login <email> <password> 			-> Backend.loginUser(email, password)
		change password <email|uid> <password> <newpassword> -> Backend.changeUserPassword(emailuid, password, newpassword)
		change email <email|uid> <password> <newemail> 		-> Backend.changeUserEmail(emailuid, password, newemail)
		user groups <email,uid>			-> Backend.userGroups(email, uid)
		delete user <email,uid>			-> Backend.deleteUser(email, uid)
		users								-> Backend.users()
		group <name> 						-> Backend.group(name)
		add <email|uid> to <group|gid>		-> Backend.addUserToGroup(emailuid, groupgid)
		remove <email|uid> from <group|gid> -> Backend.removeUserFromGroup(emailuid, groupgid)
		delete group <group|gid>			-> Backend.deleteGroup(groupgid)
		groups								-> Backend.groups()
		group users <group|gid>			-> Backend.groupUsers(groupgid)
	*/

	if strings.HasPrefix(line, "client auth ") {
		clientAuth(context, line[12:])
	} else if line == "quit" {
		quit(context, line)
	} else if line == "stats" {
		context.Ok()
	} else {
		context.Err("EFAULT")
	}
}

func ConnectionHandler(context *ConnectionContext) {
	context.Realm()
	for !context.quitting {
		line, err := context.reader.ReadString('\n')
		if err != nil {
			break // quit connection
		} else {
			line = strings.ToLower(strings.Trim(line, " \r\n"))
		}
		interpret(context, line)
	}
	context.Close()
}

func main() {
	listener, _ := net.Listen("tcp", "0.0.0.0:7654")
	fmt.Printf("ustackd listenting on 0.0.0.0:7654\n")
	var backend Backends.Abstract
	backend = new(Backends.NilBackend)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go ConnectionHandler(NewConnectionContext(conn, &backend))
	}
}
