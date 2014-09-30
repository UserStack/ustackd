package server

import (
	"fmt"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/signal"

	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/client"
	"github.com/codegangsta/cli"
)

type Server struct {
	Logger    *log.Logger
	Cfg       *Config
	Backend   backends.Abstract
	App       *cli.App
	running   bool
	listeners []net.Listener
	Stats
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

func (s *Server) Run(args []string) {
	s.App.Action = func(c *cli.Context) {
		s.RunContext(c)
	}
	s.App.Run(args)
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

	if err = s.setupBackend(); err != nil {
		logger.Printf("Setup Backend: %s\n", err)
		return
	}

	connChan := make(chan net.Conn)
	if err = s.setupListeners(connChan); err != nil {
		logger.Printf("Setup Listeners: %s\n", err)
		return
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	for s.running {
		select {
		case <-sigChan:
			s.Stop()
		case conn := <-connChan:
			go NewContext(conn, s).Handle()
		}
	}
	logger.Println("Shutdown server")
}

func (s *Server) Stop() (err error) {
	os.Remove(s.Cfg.Daemon.Pid)
	s.running = false
	for _, listener := range s.listeners {
		if err = listener.Close(); err != nil {
			return
		}
	}
	return
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
	case "mysql":
		err = s.setupMysql()
	case "postgres":
		err = s.setupPostgres()
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

func (s *Server) setupMysql() (err error) {
	mysql, err := backends.NewMysqlBackend(s.Cfg.Mysql.Url)
	if err != nil {
		err = fmt.Errorf("Unable to open mysql at %s: %s\n", s.Cfg.Mysql.Url, err)
	}
	s.Backend = &mysql
	return
}

func (s *Server) setupPostgres() (err error) {
	postgres, err := backends.NewPostgresBackend(s.Cfg.Postgres.Url)
	if err != nil {
		err = fmt.Errorf("Unable to open postgres at %s: %s\n", s.Cfg.Postgres.Url, err)
	}
	s.Backend = &postgres
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

func (s *Server) setupListeners(connChan chan net.Conn) (err error) {
	s.listeners = make([]net.Listener, len(s.Cfg.Daemon.Listen))
	for i, bindAddress := range s.Cfg.Daemon.Listen {
		listener, lerr := net.Listen("tcp", bindAddress)
		if lerr != nil {
			err = fmt.Errorf("Unable to listen: %s\n", lerr)
			return
		}
		s.Logger.Printf("ustackd listenting on " + bindAddress + "\n")

		go (func() {
			for {
				conn, err := listener.Accept()
				if err != nil && s.running {
					s.Logger.Printf("Can't accept connection: %s\n", err)
					continue
				}
				connChan <- conn
			}
		})()

		s.listeners[i] = listener
	}
	return
}
