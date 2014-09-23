package config

import (
	"testing"
	"reflect"
)

func TestRead(t *testing.T) {
	cfg, err := Read("ustack.conf")

	if err != nil {
		t.Errorf("Failed to parse gcfg data: %s", err)
	}

	expected := Config{
		Daemon{[]string {"::1:7654", "127.0.0.1:7654"}, "ustackd $VERSION$", "sqlite"},
		Syslog{"Debug"},
		Ssl{true},
		Sqlite{"ustack.db"},
	}

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("Config is expected to be %s, but is %s", expected, cfg)
	}

}

func TestNoFile(t *testing.T) {
	_, err := Read("bla.conf")

	if err == nil {
		t.Errorf("Failed to fail for non-existent file")
	}
	expectedFileNotFoundError := "open bla.conf: no such file or directory"
	if err.Error() != expectedFileNotFoundError {
		t.Errorf("Got error: %s, but expected %s, ", err.Error(), expectedFileNotFoundError)
	}
}
