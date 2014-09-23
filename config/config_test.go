package config

import (
	"testing"
)

func TestRead(t *testing.T) {
	var cfg Config
	var err error
	cfg, err = Read("ustack.conf")
	
	if err != nil {
		t.Errorf("Failed to parse gcfg data: %s", err)
	}

	var expected Config
	expected = Config {
		Daemon { "0.0.0.0", "ustackd $VERSION$", "sqlite", 7654 },
		Syslog { "Debug" },
		Ssl { true },
		Sqlite { "ustack.db" },
	}

	if cfg != expected {
		t.Errorf("Config is expected to be %s, but is %s", expected, cfg)
	}
	
}

func TestNoFile(t *testing.T) {
	var err error
	_, err = Read("bla.conf")
	
	if err == nil {
		t.Errorf("Failed to fail for non-existent file")
	}
	expectedFileNotFoundError := "open bla.conf: no such file or directory"
	if err.Error() != expectedFileNotFoundError {
		t.Errorf("Got Error: %s, but found %s, ", err.Error(), expectedFileNotFoundError)
	}	
}
