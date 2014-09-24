package server

import (
	"github.com/UserStack/ustackd/backends"
	"github.com/UserStack/ustackd/config"
	"log"
)

type Server struct {
	Logger  *log.Logger
	Cfg     *config.Config
	Backend backends.Abstract
}
