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

	var nilSlice []string
	var nilString string
	var nilInt int

	expected := Config{
		Daemon{[]string {"0.0.0.0:7654"}, "ustackd $VERSION$", "sqlite", false},
		Syslog{3, "Debug"},
		Security{nilString, nilString, nilString, nilString, nilString},
		Ssl{true, nilSlice, nilString, nilString, nilString, nilInt, nilInt},
		Sqlite{"ustack.db"},
	}	

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("Config is expected to be %s, but is %s", expected, cfg)
	}

}

func TestReadAll(t *testing.T) {
	cfg, err := Read("all_options.conf")

	if err != nil {
		t.Errorf("Failed to parse gcfg data: %s", err)
	}


	expected := Config{
		Daemon{[]string {"0.0.0.0:1234", "127.0.0.1:7654"}, "ustackd $VERSION$", "sqlite", true},
		Syslog{3, "Debug"},
		Security{
			"42421da75756d69832de50c3ab34f68ab5118b53", 
			"6d95e4ac638daf4b786e94f30dc5bf6bb7118386", 
			"/var/run/ustackd", 
			"ustack", 
			"ustack",
		},
		Ssl{
			true, 
			[]string{"::1:8765"}, 
			"/etc/ustack/key.pem", 
			"/etc/ustack/cert.pem",
			"ECDH+AESGCM:DH+AESGCM:ECDH+AES256:DH+AES256:ECDH+AES128:DH+AES:ECDH+3DES:DH+3DES:RSA+AESGCM:RSA+AES:RSA+3DES:!aNULL:!MD5:!DSS",
			303,
			303,
		},
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
