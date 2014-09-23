package main

import (
    "fmt"
    "bufio"
    "net"
    "strings"
)

// Main structs

type User struct {
    uid int
    email string
}

type Group struct {
    gid int
    name string
}

// Client

type ConnectionContext struct {
    conn net.Conn
    reader *bufio.Reader
    writer *bufio.Writer
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

func (context *ConnectionContext) Err(code string) {
    context.Write("- " + code)
}

func (context *ConnectionContext) Log(line string) {
    fmt.Printf("%s: %s\r\n", context.conn.RemoteAddr(), line)
}

func login(context *ConnectionContext, line string) {
    passwd := line[6:]
    context.Log("Try login with '" + passwd + "'")

    if passwd == "secret" {
        context.loggedin = true
        context.Ok()
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
    if strings.HasPrefix(line, "login") && len(line) >=6 {
        login(context, line)
    } else if line == "quit" {
        quit(context, line)
    } else {
        context.Err("EFAULT")
    }
}

func ConnectionHandler(conn net.Conn) {
    realm := "ustackd 0.0.1"
    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)
    context := ConnectionContext{conn, reader, writer, false, false}

    context.Log("new client connected")
    context.Write(realm + " (user group)\r\n")
    for !context.quitting {
        line, err := reader.ReadString('\n')
        if err != nil {
            break // quit connection
        } else {
            line = strings.ToLower(strings.Trim(line, " \r\n"))
        }
        interpret(&context, line)
    }
    conn.Close()
    context.Log("Client disonnected")
}

func main() {
    listener, _ := net.Listen("tcp", "0.0.0.0:7654")
    fmt.Printf("ustackd listenting on 0.0.0.0:7654\n")

    for {
        conn, err := listener.Accept()
        if err != nil {
            // handle error
            continue
        }
        go ConnectionHandler(conn)
    }
}
