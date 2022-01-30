package core

import "time"

type Speed struct {
	Download  float64
	Upload    float64
	Ping      time.Duration
	Timestamp time.Time
}

var InvalidSpeed = Speed{-1, -1, -1, time.Unix(0, 0)}

func (s *Speed) Equals(other *Speed) bool {
	return s.Download == other.Download &&
		s.Upload == other.Upload &&
		s.Ping == other.Ping &&
		s.Timestamp.Equal(other.Timestamp)
}
