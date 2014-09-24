package main

import (
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"

	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/config"
	"github.com/UserStack/ustackd/connection"
	"github.com/codegangsta/cli"
)

func main() {
	var cfg config.Config
	cfg, _ = config.Read("config/ustack.conf")

	app := cli.NewApp()
	app.Name = "ustackd"
	app.Usage = "the UserStack daemon"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "config/ustack.conf",
			Usage: "the path of the main configuration file",
		},
		cli.BoolFlag{
			Name:  "foreground, f",
			Usage: "if the app should run in foreground or not",
		},
	}
	app.Action = func(c *cli.Context) {
		bindAddress := cfg.Daemon.Listen[0]
		listener, err := net.Listen("tcp", bindAddress)
		var logger *log.Logger

		if c.Bool("foreground") || cfg.Daemon.Foreground {
			logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
		} else {
			var err error
			logger, err = syslog.NewLogger(syslog.LOG_EMERG|syslog.LOG_KERN,
				log.LstdFlags|log.Lmicroseconds)

			if err != nil {
				fmt.Printf("Unable to connecto to syslog: %s\n", err)
				return
			}
		}

		if err != nil {
			logger.Printf("Unable to listen: %s\n", err)
			return
		}

		logger.Printf("ustackd listenting on " + bindAddress + "\n")
		var backend backends.Abstract
		backend = new(backends.NilBackend)

		for {
			conn, err := listener.Accept()
			if err != nil {
				logger.Printf("Can't accept connection: %s\n", err)
				continue
			}
			go connection.NewContext(conn, logger, backend).Handle()
		}
	}

	app.Run(os.Args)
}
