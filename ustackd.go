package main

import (
	"os"

	"github.com/UserStack/ustackd/server"
)

func main() {
	server.NewServer().Run(os.Args)
}
