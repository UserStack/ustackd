package connection

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/UserStack/ustackd/backends"
)

type Context struct {
	Conn     net.Conn
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	Backend  backends.Abstract
	Loggedin bool
	Quitting bool
}

func (context *Context) Write(line string) {
	context.Writer.WriteString(line + "\r\n")
	context.Writer.Flush()
}

func (context *Context) Ok() {
	context.Write("+ OK")
}

func (context *Context) OkValue(value string) {
	context.Write("+ OK " + value)
}

func (context *Context) Err(code string) {
	context.Write("- " + code)
}

func (context *Context) Log(line string) {
	fmt.Printf("%s: %s\r\n", context.Conn.RemoteAddr(), line)
}

func (context *Context) Realm() {
	realm := "ustackd 0.0.1"
	context.Log("new client connected")
	context.Write(realm + " (user group)")
}

func (context *Context) Close() {
	context.Conn.Close()
	context.Log("Client disonnected")
}

func (context *Context) Handle() {
	context.Realm()
	for !context.Quitting {
		line, err := context.Reader.ReadString('\n')
		if err != nil {
			break // quit connection
		} else {
			line = strings.ToLower(strings.Trim(line, " \r\n"))
		}
		interpret(context, line)
	}
	context.Close()
}

func NewContext(conn net.Conn, backend backends.Abstract) *Context {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Context{conn, reader, writer, backend, false, false}
}
