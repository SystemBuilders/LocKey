package session

import (
	"github.com/SystemBuilders/LocKey/internal/lockclient/id"
)

var _ Session = (*SimpleSession)(nil)

// SimpleSession implements a session.
type SimpleSession struct {
	sessionID id.ID
	clientID  id.ID
	processID id.ID
}

// SessionID returns the sessionID of the SimpleSession.
func (s *SimpleSession) SessionID() id.ID {
	return s.sessionID
}

// ClientID returns the clientID of the SimpleSession.
func (s *SimpleSession) ClientID() id.ID {
	return s.clientID
}

// ProcessID returns the processID of the SimpleSession
func (s *SimpleSession) ProcessID() id.ID {
	return s.processID
}

// NewSession returns a new instance of a session with the given parameters.
func NewSession(sessionID, clientID, processID id.ID) Session {
	return &SimpleSession{
		sessionID: sessionID,
		clientID:  clientID,
		processID: processID,
	}
}
