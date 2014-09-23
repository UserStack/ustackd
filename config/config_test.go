package config

import (
	"testing"
	"fmt"
)

func TestRead(t *testing.T) {
	var cfg Config
	var err error
	cfg, err = read("ustack.conf")
	
	if err != nil {
		t.Errorf("Failed to parse gcfg data: %s", err)
	}
	if cfg.Daemon.Interfaces != "0.0.0.0" {
		t.Errorf("Name expected to be 'value' but is %s", cfg.Daemon.Interfaces)
	}
	fmt.Println(cfg.Daemon.Interfaces)
	
}
