package connection

import (
	"bufio"
	"log"
	"net"
	"strings"

	"github.com/UserStack/ustackd/backends"
)

type Context struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	backend  backends.Abstract
	logger   *log.Logger
	loggedin bool
	quitting bool
}

func NewContext(conn net.Conn, logger *log.Logger, backend backends.Abstract) *Context {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &Context{
		conn:    conn,
		reader:  reader,
		writer:  writer,
		backend: backend,
		logger:  logger,
	}
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
	context.logger.Printf("%s: %s\r\n", context.conn.RemoteAddr(), line)
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
	interpreter := Interpreter{context, context.backend}

	for !context.quitting {
		line, err := context.reader.ReadString('\n')
		if err != nil {
			break // quit connection
		}
		line = strings.ToLower(strings.Trim(line, " \r\n"))
		interpreter.parse(line)
	}
}
