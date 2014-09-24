package main

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/signal"
	"strconv"

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

		pidFile := cfg.Daemon.Pid_Path + "/" + app.Name + ".pid"
		checkPidFile(pidFile)
		setPidFile(pidFile)

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

		server := server.Server{logger, &cfg, backend}
		running := true
		go checkSignal(pidFile, &running, listener)

		for running {
			conn, err := listener.Accept()
			if err != nil && running {
				logger.Printf("Can't accept connection: %s\n", err)
				continue
			}
			go connection.NewContext(conn, &server).Handle()
		}
		logger.Println("Shutdown server")
	}

	app.Run(os.Args)
}

func checkSignal(pidfile string, running *bool, listener net.Listener) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c //Block until a signal is received
	os.Remove(pidfile)
	*running = false
	listener.Close()
}

func checkPidFile(filename string) {
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("Found PID: %d\n", getPidFile(filename))
	}
}

func setPidFile(filename string) {
	f, _ := os.Create(filename)
	defer f.Close()
	pid := os.Getpid()
	f.WriteString(strconv.Itoa(pid))
	fmt.Printf("Write PID: %d\n", pid)
}

func getPidFile(filename string) (pid int) {
	f, _ := os.Open(filename)
	defer f.Close()
	r := bufio.NewReaderSize(f, 5)
	line, _, _ := r.ReadLine()
	pid, _ = strconv.Atoi(string(line))
	return pid
}
