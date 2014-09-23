package config
import (
	"fmt"
	"code.google.com/p/gcfg"
)

type Config struct {
       Daemon
       Syslog
       Ssl
       Sqlite
}

type Daemon struct {
        Interfaces, Realm, Backend string
        Port int
}

type Syslog struct {
        Level string
}

type Ssl struct {
        Enabled bool        
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