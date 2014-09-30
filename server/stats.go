package server

type Stats struct {
	Connects, Disconnects, Login, FailedLogin int
}

func (s *Stats) Reset() {
	s.Connects = 0
	s.Disconnects = 0
	s.Login = 0
	s.FailedLogin = 0
}

func (s *Stats) ActiveConnections() int {
	return s.Connects - s.Disconnects
}
