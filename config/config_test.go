package config

import (
	"reflect"
	"testing"
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
		Daemon{[]string{"0.0.0.0:7654"}, "ustackd $VERSION$", "sqlite", "./", false},
		Syslog{3, "Debug"},
		ClientAuth{[]Auth{}},
		Security{nilString, nilString},
		Ssl{true, nilSlice, nilString, nilString, nilString, nilInt, nilInt},
		Sqlite{"./ustack.db"},
		Proxy{""},
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
		Daemon{[]string{"0.0.0.0:1234", "127.0.0.1:7654"}, "ustackd $VERSION$", "sqlite", "/var/run", true},
		Syslog{3, "Debug"},
		ClientAuth{[]Auth{
			Auth{"42421da75756d69832d", "//", false},
			Auth{"6d95e4ac638daf4b786", "/^(login|set|get|change (password|email))/", true},
			Auth{"04d6eb93ab5d30f7bb0", "/^(users|groups|group users)/", false},
		},
		},
		Security{
			"/var/run/ustackd",
			"ustack",
		},
		Ssl{
			true,
			[]string{"::1:8765"},
			"/etc/ustack/key.pem",
			"/etc/ustack/cert.pem",
			"ECDH+AESGCM:DH+AESGCM:ECDH+AES256:DH+AES256:ECDH+AES128:DH+AES:ECDH+3DES:DH+3DES:RSA+AESGCM:RSA+AES:RSA+3DES:!aNULL:!MD5:!DSS",
			771,
			771,
		},
		Sqlite{"ustack.db"},
		Proxy{"127.0.0.1:7654"},
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

func TestSplitAuth(t *testing.T) {
	var configIntern ConfigIntern
	configIntern.Client = Client{Auth: []string{"a:b:c"}}
	config, err := splitAuth(configIntern)
	if err != nil {
		t.Error(err.Error())
	}

	expected := ClientAuth{[]Auth{Auth{"a", "c", false}}}
	if !reflect.DeepEqual(config.ClientAuth, expected) {
		t.Errorf("Config.ClientAuth is expected to be %+v, but is %+v", expected, config.ClientAuth)
	}

}

func TestSplitAuthAllow(t *testing.T) {
	var configIntern ConfigIntern
	configIntern.Client = Client{Auth: []string{"a:allow:c"}}
	config, err := splitAuth(configIntern)
	if err != nil {
		t.Error(err.Error())
	}

	expected := ClientAuth{[]Auth{Auth{"a", "c", true}}}
	if !reflect.DeepEqual(config.ClientAuth, expected) {
		t.Errorf("Config.ClientAuth is expected to be %+v, but is %+v", expected, config.ClientAuth)
	}

}

func TestSplitAuthFail(t *testing.T) {
	var configIntern ConfigIntern
	configIntern.Client = Client{Auth: []string{"a:b"}}
	_, err := splitAuth(configIntern)
	if err.Error() != "Could not split [client] auth line into 3 parts (Id, Command, Regex): a:b" {
		t.Error("Failed to fail when 1 part is missing")
	}

}
