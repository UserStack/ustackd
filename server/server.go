package server

import (
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"

	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/client"
	"github.com/codegangsta/cli"
)

type Server struct {
	Logger  *log.Logger
	Cfg     *Config
	Backend backends.Abstract
	App     *cli.App
}

func NewServer(app *cli.App) *Server {
	return &Server{App: app}
}

func (server *Server) Run(c *cli.Context) {
	cfg, err := Read(c.String("config"))
	if err != nil {
		fmt.Printf("Unable read config file: %s\n", err)
		return
	}
	server.Cfg = &cfg

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
	server.Logger = logger

	if err = server.Demonize(); err != nil {
		logger.Printf("Unable to demonize: %s\n", err)
		return
	}

	bindAddress := cfg.Daemon.Listen[0]
	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		logger.Printf("Unable to listen: %s\n", err)
		return
	}
	logger.Printf("ustackd listenting on " + bindAddress + "\n")

	// get the right backend for the configuration
	switch cfg.Daemon.Backend {
	case "sqlite":
		sqlite, serr := backends.NewSqliteBackend(cfg.Sqlite.Url)
		if serr != nil {
			logger.Printf("Unable to open sqlite at %s: %s\n",
				cfg.Sqlite.Url, serr)
			return
		}
		server.Backend = &sqlite
	case "proxy":
		proxy, serr := client.Dial(cfg.Proxy.Host)
		if serr != nil {
			logger.Printf("Unable to connect to %s: %s\n",
				cfg.Proxy.Host, serr)
			return
		}
		if cfg.Proxy.Ssl {
			if len(cfg.Proxy.Cert) > 0 {
				serr = proxy.StartTlsWithoutCertCheck()
			} else {
				serr = proxy.StartTlsWithCert(cfg.Proxy.Cert)
			}
		}
		if serr != nil {
			logger.Printf("Unable to open proxy for %s: %s\n",
				cfg.Proxy.Host, serr)
			return
		}
		if len(cfg.Proxy.Passwd) > 0 { // if passwd given
			aerr := proxy.ClientAuth(cfg.Proxy.Passwd)
			if aerr != nil {
				logger.Printf("Unable to authenticate with %s: %s\n",
					cfg.Proxy.Host, aerr.Code)
				return
			}
		}
		server.Backend = proxy
	case "nil":
		server.Backend = &backends.NilBackend{}
	default:
		logger.Printf("Unkown backend: %s\n", cfg.Daemon.Backend)
		return
	}

	isRunning := true
	go server.CheckSignal(&isRunning, listener.Close)

	for isRunning {
		conn, err := listener.Accept()
		if err != nil && isRunning {
			logger.Printf("Can't accept connection: %s\n", err)
			continue
		}
		go NewContext(conn, server).Handle()
	}
	logger.Println("Shutdown server")
}
