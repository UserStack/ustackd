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
	Logger   *log.Logger
	Cfg      *Config
	Backend  backends.Abstract
	App      *cli.App
	running  bool
	listener net.Listener
}

func NewServer() *Server {
	app := cli.NewApp()
	app.Name = "ustackd"
	app.Usage = "the UserStack daemon"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "config/ustackd.conf",
			Usage: "the path of the main configuration file",
		},
		cli.BoolFlag{
			Name:  "foreground, f",
			Usage: "if the app should run in foreground or not",
		},
	}
	return &Server{App: app, running: true}
}

func (server *Server) Run(args []string) {
	server.App.Action = func(c *cli.Context) {
		server.RunContext(c)
	}
	server.App.Run(os.Args)
}

func (s *Server) RunContext(c *cli.Context) {
	cfg, err := Read(c.String("config"))
	if err != nil {
		fmt.Printf("Unable read config file: %s\n", err)
		return
	}
	s.Cfg = &cfg

	if c.Bool("foreground") {
		cfg.Daemon.Foreground = true
	}

	logger, err := s.setupLogger()
	if err != nil {
		fmt.Printf("Unable to connect to syslog: %s\n", err)
		return
	}

	if err = s.demonize(); err != nil {
		logger.Printf("Unable to demonize: %s\n", err)
		return
	}

	bindAddress := cfg.Daemon.Listen[0]
	server.listener, err = net.Listen("tcp", bindAddress)
	if err != nil {
		logger.Printf("Unable to listen: %s\n", err)
		return
	}
	logger.Printf("ustackd listenting on " + bindAddress + "\n")

	if err = s.setupBackend(); err != nil {
		logger.Printf("Setup Backend: %s\n", err)
		return
	}

	go s.checkSignal(server.Stop)
	for server.running {
		conn, err := server.listener.Accept()
		if err != nil && server.running {
			logger.Printf("Can't accept connection: %s\n", err)
			continue
		}
		go NewContext(conn, s).Handle()
	}
	logger.Println("Shutdown server")
}

func (server *Server) Stop() error {
	server.running = false
	return server.listener.Close()
}

func (s *Server) setupLogger() (logger *log.Logger, err error) {

	flags := log.LstdFlags | log.Lmicroseconds
	if s.Cfg.Daemon.Foreground {
		logger = log.New(os.Stdout, "", flags)
	} else {
		logger, err = syslog.NewLogger(s.Cfg.Syslog.Severity|s.Cfg.Syslog.Facility, flags)
	}
	s.Logger = logger
	return
}

func (s *Server) setupBackend() (err error) {
	// get the right backend for the configuration
	switch s.Cfg.Daemon.Backend {
	case "sqlite":
		err = s.setupSqlite()
	case "proxy":
		err = s.setupProxy()
	case "nil":
		s.Backend = &backends.NilBackend{}
	default:
		err = fmt.Errorf("Unknown backend: %s\n", s.Cfg.Daemon.Backend)
	}
	return
}

func (s *Server) setupSqlite() (err error) {
	sqlite, err := backends.NewSqliteBackend(s.Cfg.Sqlite.Url)
	if err != nil {
		err = fmt.Errorf("Unable to open sqlite at %s: %s\n", s.Cfg.Sqlite.Url, err)
	}
	s.Backend = &sqlite
	return
}

func (s *Server) setupProxy() (err error) {
	Proxy := s.Cfg.Proxy
	proxy, err := client.Dial(Proxy.Host)
	if err != nil {
		err = fmt.Errorf("Unable to connect to %s: %s\n", Proxy.Host, err)
		return
	}
	if Proxy.Ssl {
		if len(Proxy.Cert) > 0 {
			err = proxy.StartTlsWithoutCertCheck()
		} else {
			err = proxy.StartTlsWithCert(Proxy.Cert)
		}
		if err != nil {
			err = fmt.Errorf("Unable to open proxy for %s: %s\n", Proxy.Host, err)
			return
		}
	}

	if len(Proxy.Passwd) > 0 { // if passwd given
		if perr := proxy.ClientAuth(Proxy.Passwd); perr != nil {
			err = fmt.Errorf("Unable to authenticate with %s: %s\n", Proxy.Host, perr.Code)
		}
	}
	s.Backend = proxy
	return
}
