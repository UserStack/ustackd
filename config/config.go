package config
import (
	"fmt"
	"code.google.com/p/gcfg"
)

type Config struct {
        Daemon struct {
                Interfaces, Realm, Backend string
                Port int
        }
        Syslog struct {
        	Level string
        }
        Ssl struct {
                Enabled bool
        	
        }
        Sqlite struct {
        	Url string
        }

}

func Read(filename string) (Config, error) {	
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, filename)
	if err != nil {
		fmt.Printf("Failed to parse %s: %s\n", filename, err)
	}
	return cfg, err
}
