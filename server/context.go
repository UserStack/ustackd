package server

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
)

type Context struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	*Server
	addr     net.Addr
	quitting bool
}

func NewContext(conn net.Conn, server *Server) *Context {
	return &Context{
		conn: conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		Server: server,
		addr: conn.RemoteAddr(),
	}
}

func (context *Context) Write(line string) {
	context.Log("<- " + line)
	context.writer.WriteString(line + "\r\n")
	context.writer.Flush()
}

func (context *Context) Writef(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	context.Log("<- " + line)
	context.writer.WriteString(line + "\r\n")
	context.writer.Flush()
}

func (context *Context) Ok() {
	context.Write("+ OK")
}

func (context *Context) OkValue(value interface{}) {
	context.Writef("+ OK %v", value)
}

func (context *Context) Err(code string) {
	context.Write("- " + code)
}

func (context *Context) Log(line string) {
	context.Logger.Printf("%s: %s\r\n", context.addr, line)
}

func (context *Context) Realm() {
	realm := strings.Replace(context.Cfg.Daemon.Realm, "$VERSION$",
		context.App.Version, 1)
	context.Write(realm)
	context.Log("Client connected")
	context.Server.Stats.Connects++
}

func (context *Context) Close() {
	context.Log("Client disconnected")
	context.Server.Stats.Disconnects++
	context.conn.Close()
}

func (context *Context) Handle() {
	context.Realm()
	defer context.Close()
	interpreter := Interpreter{Context: context}

	for !context.quitting {
		line, err := context.reader.ReadString('\n')
		if err != nil {
			break // quit connection
		}
		line = strings.Trim(line, " \r\n")
		context.Log("-> " + line)
		if context.starttls(line) {
			continue
		}
		interpreter.parse(line)
	}
}

func (context *Context) starttls(line string) bool {
	if line == "starttls" && context.tlsConfig != nil {
		context.conn = tls.Server(context.conn, context.tlsConfig)
		context.reader = bufio.NewReader(context.conn)
		context.writer = bufio.NewWriter(context.conn)
		context.Log("Changed to tls channel")
		return true
	} else {
		return false
	}
}
