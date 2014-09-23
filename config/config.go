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
        Logging struct {
        	Level string
        	Facility int
        }
        Ssl struct {
        	
        }
        Sqlite struct {
        	Url string
        }

}

func read(filename string) (Config, error) {	
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, filename)
	if err != nil {
		fmt.Printf("Failed to parse gcfg data: %s", err)
	}
	return cfg, err
}
