package server

import (
	"log/syslog"
	"reflect"
	"testing"
)

func TestRead(t *testing.T) {
	cfg, err := Read("../config/ustackd.conf")

	if err != nil {
		t.Errorf("Failed to parse gcfg data: %s", err)
	}

	var nilString string

	expected := Config{
		Daemon{[]string{"0.0.0.0:7654"}, "ustackd $VERSION$", "sqlite", "./ustackd.pid", false},
		Syslog{syslog.LOG_FTP, syslog.LOG_DEBUG},
		Client{[]Auth{}},
		Security{nilString, nilString},

		Ssl{true, "config/key.pem", "config/cert.pem"},
		Sqlite{"./ustackd.db"},
		Proxy{"", false, "", ""},
	}

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("Config is expected to be %v, but is %v", expected, cfg)
	}

}

func TestDefault(t *testing.T) {
	cfg, err := Read("../config/no.conf")

	if err != nil {
		t.Errorf("Failed to parse gcfg data: %s", err)
	}

	var nilString string

	expected := Config{
		Daemon{[]string{"0.0.0.0:7654"}, "ustackd $VERSION$", "nil", "./ustackd.pid", true},
		Syslog{syslog.LOG_DAEMON, syslog.LOG_EMERG},
		Client{[]Auth{}},
		Security{nilString, nilString},

		Ssl{false, nilString, nilString},
		Sqlite{nilString},
		Proxy{nilString, false, nilString, nilString},
	}

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("Config is expected to be %v, but is %v", expected, cfg)
	}
}

func TestReadAll(t *testing.T) {
	cfg, err := Read("../config/all_options.conf")

	if err != nil {
		t.Errorf("Failed to parse gcfg data: %s", err)
	}

	expected := Config{
		Daemon{[]string{"0.0.0.0:1234", "127.0.0.1:7654"}, "ustackd $VERSION$", "sqlite", "/var/run/ustackd.pid", true},
		Syslog{syslog.LOG_FTP, syslog.LOG_DEBUG},
		Client{[]Auth{
			Auth{"42421da75756d69832d", ".*", true},
			Auth{"6d95e4ac638daf4b786", "^(login|set|get|change (password|email))", true},
			Auth{"04d6eb93ab5d30f7bb0", "^(users|groups|group users)", false},
		},
		},
		Security{
			"/var/run/ustackd",
			"ustack",
		},
		Ssl{
			true,
			"/etc/ustack/key.pem",
			"/etc/ustack/cert.pem",
		},
		Sqlite{"ustackd.db"},
		Proxy{"127.0.0.1:7654", true, "config/cert.pem", "SOMEVERYGOODSECRET"},
	}

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("Config is expected to be %v, but is %v", expected, cfg)
	}
}

func TestNoFile(t *testing.T) {
	_, err := Read("../config/bla.conf")

	if err == nil {
		t.Errorf("Failed to fail for non-existent file")
	}
	expectedFileNotFoundError := "open ../config/bla.conf: no such file or directory"
	if err.Error() != expectedFileNotFoundError {
		t.Errorf("Got error: %v, but expected %v, ", err.Error(), expectedFileNotFoundError)
	}
}

func TestSplitAuth(t *testing.T) {
	clientIntern := ClientIntern{Auth: []string{"a:b:c"}}
	client, err := splitAuth(clientIntern)
	if err != nil {
		t.Error(err.Error())
	}

	expected := Client{[]Auth{Auth{"a", "c", false}}}
	if !reflect.DeepEqual(client, expected) {
		t.Errorf("Config.Client is expected to be %+v, but is %+v", expected, client)
	}

}

func TestSplitAuthAllow(t *testing.T) {
	clientIntern := ClientIntern{Auth: []string{"a:allow:c"}}
	client, err := splitAuth(clientIntern)
	if err != nil {
		t.Error(err.Error())
	}

	expected := Client{[]Auth{Auth{"a", "c", true}}}
	if !reflect.DeepEqual(client, expected) {
		t.Errorf("Config.Client is expected to be %+v, but is %+v", expected, client)
	}

}

func TestSplitAuthFail(t *testing.T) {
	clientIntern := ClientIntern{Auth: []string{"a:b"}}
	_, err := splitAuth(clientIntern)
	if err.Error() != "Could not split [client] auth line into 3 parts (Passwd, Command, Regex): a:b" {
		t.Error("Failed to fail when 1 part is missing")
	}

}
