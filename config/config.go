package config

import (
	"code.google.com/p/gcfg"
	"fmt"
)

type Config struct {
	Daemon
	Syslog
	Security
	Ssl
	Sqlite
}

type Daemon struct {
	Listen []string
	Realm, Backend string
	Foreground bool
}

type Syslog struct {
	Level string
}
type Security struct {
	Secret, Admin_Secret, Chroot, Gid, Uid string
}

type Ssl struct {
	Enabled bool
	Listen []string
	Key, Cert, Protocols, Ciphers string
}

type Sqlite struct {
	Url string
}

func Read(filename string) (Config, error) {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, filename)
	if err != nil {
		fmt.Printf("Failed to parse %s: %s\n", filename, err)
	}
	return cfg, err
}
