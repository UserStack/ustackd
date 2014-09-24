package server

import (
	"log"
	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/config"
)

type Server struct {
	Logger   *log.Logger
	Cfg      *config.Config
	Backend  backends.Abstract
}
