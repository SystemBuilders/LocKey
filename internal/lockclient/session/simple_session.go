package session

import (
	"github.com/oklog/ulid"
)

var _ Session = (*SimpleSession)(nil)

// SimpleSession implements a session.
type SimpleSession struct {
	sessionID ulid.ULID
	clientID  ulid.ULID
	processID ulid.ULID
}

// SessionID returns the sessionID of the SimpleSession.
func (s *SimpleSession) SessionID() ulid.ULID {
	return s.sessionID
}

// ClientID returns the clientID of the SimpleSession.
func (s *SimpleSession) ClientID() ulid.ULID {
	return s.clientID
}

// ProcessID returns the processID of the SimpleSession
func (s *SimpleSession) ProcessID() ulid.ULID {
	return s.processID
}

func NewSession(sessionID, clientID, processID ulid.ULID) Session {
	return &SimpleSession{
		sessionID: sessionID,
		clientID:  clientID,
		processID: processID,
	}
}
