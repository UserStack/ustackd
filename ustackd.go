package main

import (
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/signal"

	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/config"
	"github.com/UserStack/ustackd/connection"
	"github.com/UserStack/ustackd/server"
	"github.com/codegangsta/cli"
)

func main() {
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
		cfg, err := config.Read(c.String("config"))
		if err != nil {
			fmt.Printf("Unable read config file: %s\n", err)
			return
		}

		var logger *log.Logger
		if c.Bool("foreground") || cfg.Daemon.Foreground {
			logger = log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
		} else {
			var err error
			logger, err = syslog.NewLogger(syslog.LOG_EMERG|syslog.LOG_KERN,
				log.LstdFlags|log.Lmicroseconds)

			if err != nil {
				fmt.Printf("Unable to connect to syslog: %s\n", err)
				return
			}
		}
		var backend backends.Abstract
		server := server.Server{logger, &cfg, backend, app.Name}
		if err = server.Demonize(); err != nil {
			return
		}

		bindAddress := cfg.Daemon.Listen[0]
		listener, err := net.Listen("tcp", bindAddress)
		if err != nil {
			logger.Printf("Unable to listen: %s\n", err)
			return
		}
		logger.Printf("ustackd listenting on " + bindAddress + "\n")

		sqlite, serr := backends.NewSqliteBackend(cfg.Sqlite.Url)
		if serr != nil {
			logger.Printf("Unable to open sqlite at %s: %s\n",
				cfg.Sqlite.Url, serr)
			return
		}
		backend = &sqlite

		isRunning := true
		pidFile := cfg.Daemon.Pid_Path + "/" + app.Name + ".pid"
		go checkSignal(pidFile, &isRunning, listener)

		for isRunning {
			conn, err := listener.Accept()
			if err != nil && isRunning {
				logger.Printf("Can't accept connection: %s\n", err)
				continue
			}
			go connection.NewContext(conn, &server).Handle()
		}
		logger.Println("Shutdown server")
	}

	app.Run(os.Args)
}

func checkSignal(pidfile string, isRunning *bool, listener net.Listener) {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill)
	<-channel //Block until a signal is received
	os.Remove(pidfile)
	*isRunning = false
	listener.Close()
}
