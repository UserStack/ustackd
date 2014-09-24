package config

import (
	"code.google.com/p/gcfg"
	"fmt"
	"strings"
)

type ConfigIntern struct {
	Daemon
	Syslog
	Client
	Security
	Ssl
	Sqlite
}

type Config struct {
	Daemon
	Syslog
	ClientAuth
	Security
	Ssl
	Sqlite
}

type Daemon struct {
	Listen         []string
	Realm, Backend string
	Foreground     bool
}

type Syslog struct {
	Facility int
	Level    string
}

type Client struct {
	Auth []string
}

type ClientAuth struct {
	Auth []Auth
}

type Auth struct {
	Id, Regex string
	Allow     bool
}

type Security struct {
	Chroot, Gid, Uid string
}

type Ssl struct {
	Enabled                    bool
	Listen                     []string
	Key, Cert, Ciphers         string
	Protocol_Min, Protocol_Max int
}

type Sqlite struct {
	Url string
}

func Read(filename string) (Config, error) {
	var cfgIntern ConfigIntern
	var config Config
	err := gcfg.ReadFileInto(&cfgIntern, filename)
	if err != nil {
		fmt.Printf("Failed to parse %s: %s\n", filename, err)
		return config, err
	}

	config, err = splitAuth(cfgIntern)
	return config, err
}

func splitAuth(cfgIntern ConfigIntern) (Config, error) {
	var config Config
	config.Daemon = cfgIntern.Daemon
	config.Syslog = cfgIntern.Syslog
	config.Security = cfgIntern.Security
	config.Ssl = cfgIntern.Ssl
	config.Sqlite = cfgIntern.Sqlite

	var clientAuth ClientAuth
	clientAuth.Auth = make([]Auth, len(cfgIntern.Client.Auth))
	config.ClientAuth = clientAuth

	for i, line := range cfgIntern.Client.Auth {
		splitAuth := strings.SplitN(line, ":", 3)

		if len(splitAuth) != 3 {
			return config, fmt.Errorf("Could not split [client] auth line into 3 parts (Id, Command, Regex): %s", line)
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
		clientAuth.Auth[i] = auth
	}

	var nilError error

	return config, nilError
}
