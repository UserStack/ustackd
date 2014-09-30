package server

type Stats struct {
	Connects, Disconnects, Login, FailedLogin, unrestrictedCommands, restrictedCommands,
	restrictedCommandsAccessDenied int
}

func (s *Stats) Reset() {
	s.Connects = 0
	s.Disconnects = 0
	s.Login = 0
	s.FailedLogin = 0
	s.restrictedCommands = 0
	s.restrictedCommandsAccessDenied = 0
	s.unrestrictedCommands = 0
}

func (s *Stats) ActiveConnections() int {
	return s.Connects - s.Disconnects
}
