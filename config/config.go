package config

import (
	"code.google.com/p/gcfg"
	"fmt"
	"log/syslog"
	"strings"
)

type ConfigIntern struct {
	Daemon   Daemon
	Syslog   SyslogIntern
	Client   ClientIntern
	Security Security
	Ssl      Ssl
	Sqlite   Sqlite
	Proxy    Proxy
}

type Config struct {
	Daemon
	Syslog
	Client
	Security
	Ssl
	Sqlite
	Proxy
}

type Daemon struct {
	Listen              []string
	Realm, Backend, Pid string
	Foreground          bool
}

type SyslogIntern struct {
	Facility, Level string
}

type Syslog struct {
	Facility, Severity syslog.Priority
}

type ClientIntern struct {
	Auth []string
}

type Client struct {
	Auth []Auth
}

type Auth struct {
	Id, Regex string
	Allow     bool
}

type Security struct {
	Chroot, Uid string
}

type Ssl struct {
	Enabled   bool
	Key, Cert string
}

type Sqlite struct {
	Url string
}

type Proxy struct {
	Host string
	Ssl  bool
	Cert string
}

func Read(filename string) (config Config, err error) {
	var cfgIntern ConfigIntern
	err = gcfg.ReadFileInto(&cfgIntern, filename)
	if err != nil {
		return
	}
	setDefaultsIfRequired(&cfgIntern)

	config.Daemon = cfgIntern.Daemon

	config.Syslog, err = translateSyslog(cfgIntern.Syslog)
	if err != nil {
		fmt.Println(err)
		return
	}

	config.Security = cfgIntern.Security
	config.Ssl = cfgIntern.Ssl
	config.Sqlite = cfgIntern.Sqlite
	config.Proxy = cfgIntern.Proxy

	config.Client, err = splitAuth(cfgIntern.Client)
	return
}

func splitAuth(clientIntern ClientIntern) (client Client, err error) {
	client.Auth = make([]Auth, len(clientIntern.Auth))

	for i, line := range clientIntern.Auth {
		splitAuth := strings.SplitN(line, ":", 3)

		if len(splitAuth) != 3 {
			err = fmt.Errorf("Could not split [client] auth line into 3 parts (Id, Command, Regex): %s", line)
			return
		}

		var auth Auth
		for j, word := range splitAuth {
			if j == 0 {
				auth.Id = word
			}
			if j == 1 {
				auth.Allow = word == "allow"
			}
			if j == 2 {
				auth.Regex = word
			}
		}
		client.Auth[i] = auth
	}
	return
}

func translateSyslog(syslogIntern SyslogIntern) (sys Syslog, err error) {
	// nothing was set, use defaults
	if syslogIntern.Level == "" && syslogIntern.Facility == "" {
		return Syslog{syslog.LOG_DAEMON, syslog.LOG_EMERG}, nil
	}

	severities := map[string]syslog.Priority{
		"EMERG":   syslog.LOG_EMERG,
		"ALERT":   syslog.LOG_ALERT,
		"CRIT":    syslog.LOG_CRIT,
		"ERR":     syslog.LOG_ERR,
		"WARNING": syslog.LOG_WARNING,
		"NOTICE":  syslog.LOG_NOTICE,
		"INFO":    syslog.LOG_INFO,
		"DEBUG":   syslog.LOG_DEBUG,
	}
	var ok bool
	sys.Severity, ok = severities[syslogIntern.Level]
	if !ok {
		err = fmt.Errorf("Severity \"%s\" not found.", syslogIntern.Level)
		return
	}
	facilities := map[string]syslog.Priority{
		"KERN":     syslog.LOG_KERN,
		"USER":     syslog.LOG_USER,
		"MAIL":     syslog.LOG_MAIL,
		"DAEMON":   syslog.LOG_DAEMON,
		"AUTH":     syslog.LOG_AUTH,
		"SYSLOG":   syslog.LOG_SYSLOG,
		"LPR":      syslog.LOG_LPR,
		"NEWS":     syslog.LOG_NEWS,
		"UUCP":     syslog.LOG_UUCP,
		"CRON":     syslog.LOG_CRON,
		"AUTHPRIV": syslog.LOG_AUTHPRIV,
		"FTP":      syslog.LOG_FTP,
		"LOCAL0":   syslog.LOG_LOCAL0,
		"LOCAL1":   syslog.LOG_LOCAL1,
		"LOCAL2":   syslog.LOG_LOCAL2,
		"LOCAL3":   syslog.LOG_LOCAL3,
		"LOCAL4":   syslog.LOG_LOCAL4,
		"LOCAL5":   syslog.LOG_LOCAL5,
		"LOCAL6":   syslog.LOG_LOCAL6,
		"LOCAL7":   syslog.LOG_LOCAL7,
	}
	sys.Facility, ok = facilities[syslogIntern.Facility]
	if !ok {
		err = fmt.Errorf("Facility \"%s\" not found.", syslogIntern.Facility)
	}
	return
}

func setDefaultsIfRequired(cfg *ConfigIntern) {
	if len(cfg.Daemon.Listen) == 0 && cfg.Daemon.Realm == "" &&
		cfg.Daemon.Backend == "" && cfg.Daemon.Pid == "" &&
		cfg.Syslog.Level == "" && cfg.Syslog.Facility == "" {
		// if nothing is set assume foreground
		cfg.Daemon.Foreground = true
	}
	if len(cfg.Daemon.Listen) == 0 {
		cfg.Daemon.Listen = []string{"0.0.0.0:7654"}
	}
	if cfg.Daemon.Realm == "" {
		cfg.Daemon.Realm = "ustackd $VERSION$"
	}
	if cfg.Daemon.Backend == "" {
		cfg.Daemon.Backend = "nil"
	}
	if cfg.Daemon.Pid == "" {
		cfg.Daemon.Pid = "./ustackd.pid"
	}
}
