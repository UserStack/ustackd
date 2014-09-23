package connection

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/UserStack/ustackd/backends"
)

type Context struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	backend  backends.Abstract
	loggedin bool
	quitting bool
}

func (context *Context) Write(line string) {
	context.writer.WriteString(line + "\r\n")
	context.writer.Flush()
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
	fmt.Printf("%s: %s\r\n", context.conn.RemoteAddr(), line)
}

func (context *Context) Realm() {
	realm := "ustackd 0.0.1"
	context.Log("new client connected")
	context.Write(realm + " (user group)")
}

func (context *Context) Close() {
	context.conn.Close()
	context.Log("Client disonnected")
}

func (context *Context) Handle() {
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

func NewContext(conn net.Conn, backend backends.Abstract) *Context {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Context{conn, reader, writer, backend, false, false}
}
