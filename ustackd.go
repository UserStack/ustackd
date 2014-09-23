package main

import (
	"fmt"
	"net"

	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/config"
	"github.com/UserStack/ustackd/connection"
)

func main() {
	var cfg config.Config
	cfg, _ = config.Read("config/ustack.conf")
	bindAddress := cfg.Daemon.Listen
	listener, err := net.Listen("tcp", bindAddress)

	if err != nil {
		fmt.Printf("Unable to listen: %s\n", err)
	} else {
		fmt.Printf("ustackd listenting on " + bindAddress + "\n")
		var backend backends.Abstract
		backend = new(backends.NilBackend)

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("Can't accept connection: %s\n", err)
				continue
			}
			go connection.NewContext(conn, backend).Handle()
		}
	}
}
