package connection

import (
	"bufio"
	"net"
	"strings"

	"github.com/UserStack/ustackd/server"
)

type Context struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	*server.Server
	loggedin bool
	quitting bool
}

func NewContext(conn net.Conn, server *server.Server) *Context {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Context{conn, reader, writer, server, false, false}
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
	context.Logger.Printf("%s: %s\r\n", context.conn.RemoteAddr(), line)
}

func (context *Context) Realm() {
	realm := "ustackd 0.0.1"
	context.Log("new client connected")
	context.Write(realm + " (user group)")
}

func (context *Context) Close() {
	context.conn.Close()
	context.Log("Client disconnected")
}

func (context *Context) Handle() {
	context.Realm()
	defer context.Close()
	interpreter := Interpreter{context}

	for !context.quitting {
		line, err := context.reader.ReadString('\n')
		if err != nil {
			break // quit connection
		}
		line = strings.ToLower(strings.Trim(line, " \r\n"))
		interpreter.parse(line)
	}
}
