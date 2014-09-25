package main

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"

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

		pidFile := cfg.Daemon.Pid_Path + "/" + app.Name + ".pid"
		err = checkPidFile(pidFile, app.Name)
		if err != nil {
			logger.Println(err.Error())
			return
		}
		writePidFile(pidFile)

		bindAddress := cfg.Daemon.Listen[0]
		listener, err := net.Listen("tcp", bindAddress)

		if err != nil {
			logger.Printf("Unable to listen: %s\n", err)
			return
		}

		logger.Printf("ustackd listenting on " + bindAddress + "\n")
		var backend backends.Abstract
		sqlite, serr := backends.NewSqliteBackend(cfg.Sqlite.Url)
		if serr != nil {
			logger.Printf("Unable to open sqlite at %s: %s\n",
							cfg.Sqlite.Url, serr)
			return
		}
		backend = &sqlite

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
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill)
	<-channel //Block until a signal is received
	os.Remove(pidfile)
	*running = false
	listener.Close()
}

func checkPidFile(pidFile, appname string) (err error) {
	if _, ferr := os.Stat(pidFile); ferr != nil {
		return
	}
	pid, err := readPidFile(pidFile)
	if err != nil {
		return
	}
	output, err := exec.Command("ps", "-o", "command=", strconv.Itoa(pid)).Output()
	if err != nil {
		return
	}
	if strings.Contains(string(output), appname) {
		err = fmt.Errorf("Running %s found with PID: %d", appname, pid)
		return
	}
	os.Remove(pidFile)
	return
}

func readPidFile(pidFile string) (pid int, err error) {
	file, err := os.Open(pidFile)
	defer file.Close()
	if err != nil {
		return
	}
	reader := bufio.NewReaderSize(file, 5)
	line, _, err := reader.ReadLine()
	if err != nil {
		return
	}
	pid, err = strconv.Atoi(string(line))
	return
}

func writePidFile(pidFile string) {
	file, err := os.Create(pidFile)
	defer file.Close()
	if err == nil {
		pid := os.Getpid()
		file.WriteString(strconv.Itoa(pid))
	}
}
