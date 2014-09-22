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

func ConnectionHandler(conn net.Conn) {
    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)
    secret := "secret"
    loggedin := false
    quitting := false
    realm := "ustackd 0.0.1"

    fmt.Printf("new client connected %s\n", conn.RemoteAddr())
    writer.WriteString(realm + " (user group)\r\n")
    writer.Flush()
    for !quitting {
        line, err := reader.ReadString('\n')
        if err != nil {
            // handle error
            continue
        } else {
            line = strings.ToLower(strings.Trim(line, " \r\n"))
        }

        if strings.HasPrefix(line, "login") && len(line) >=6 {
            passed := line[6:]

            fmt.Printf("Try login with '%s'\n", passed)

            if passed == secret {
                writer.WriteString("+ OK\r\n")
                loggedin = true
            } else {
                writer.WriteString("+ EPERM\r\n")
            }
        } else if line == "quit" {
            writer.WriteString("+ BYE\r\n")
            quitting = true
        } else {
            writer.WriteString("+ EFAULT\r\n")
        }

        if loggedin == true {
            fmt.Printf("%b\n", loggedin)
        }

        fmt.Printf("%s\r\n", line)
        writer.Flush()
    }
    conn.Close()
    fmt.Printf("Client disonnected %s\n", conn.RemoteAddr())
}

func main() {
    listener, _ := net.Listen("tcp", ":7654")
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
