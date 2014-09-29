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

	if c.Bool("foreground") {
		cfg.Daemon.Foreground = true
	}

	logger, err := server.setupLogger()
	if err != nil {
		fmt.Printf("Unable to connect to syslog: %s\n", err)
		return
	}

	if err = server.demonize(); err != nil {
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

	if err = server.setupBackend(); err != nil {
		logger.Printf("Setup Backend: %s\n", err)
		return
	}

	isRunning := true
	go server.checkSignal(&isRunning, listener.Close)

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

func (server *Server) setupLogger() (logger *log.Logger, err error) {
	flags := log.LstdFlags | log.Lmicroseconds
	if server.Cfg.Daemon.Foreground {
		logger = log.New(os.Stdout, "", flags)
	} else {
		logger, err = syslog.NewLogger(server.Cfg.Syslog.Severity|server.Cfg.Syslog.Facility, flags)
	}
	server.Logger = logger
	return
}

func (server *Server) setupBackend() (err error) {
	// get the right backend for the configuration
	switch server.Cfg.Daemon.Backend {
	case "sqlite":
		err = server.setupSqlite()
	case "proxy":
		err = server.setupProxy()
	case "nil":
		server.Backend = &backends.NilBackend{}
	default:
		err = fmt.Errorf("Unkown backend: %s\n", server.Cfg.Daemon.Backend)
	}
	return
}

func (server *Server) setupSqlite() (err error) {
	sqlite, err := backends.NewSqliteBackend(server.Cfg.Sqlite.Url)
	if err != nil {
		err = fmt.Errorf("Unable to open sqlite at %s: %s\n", server.Cfg.Sqlite.Url, err)
	}
	server.Backend = &sqlite
	return
}

func (server *Server) setupProxy() (err error) {
	cfg := server.Cfg
	proxy, err := client.Dial(cfg.Proxy.Host)
	if err != nil {
		err = fmt.Errorf("Unable to connect to %s: %s\n", cfg.Proxy.Host, err)
	}
	if cfg.Proxy.Ssl {
		if len(cfg.Proxy.Cert) > 0 {
			err = proxy.StartTlsWithoutCertCheck()
		} else {
			err = proxy.StartTlsWithCert(cfg.Proxy.Cert)
		}
	}

	if err != nil {
		err = fmt.Errorf("Unable to open proxy for %s: %s\n", cfg.Proxy.Host, err)
	}
	if len(cfg.Proxy.Passwd) > 0 { // if passwd given
		if perr := proxy.ClientAuth(cfg.Proxy.Passwd); perr != nil {
			err = fmt.Errorf("Unable to authenticate with %s: %s\n", cfg.Proxy.Host, perr.Code)
		}
	}
	server.Backend = proxy
	return
}
