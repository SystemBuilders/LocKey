package lockclient

// Error provides constant error strings to the driver functions.
type Error string

func (e Error) Error() string { return string(e) }

// Constant errors.
// Rule of thumb, all errors start with a small letter and end with no full stop.
const (
	ErrSessionInexistent = Error("the session related to this process doesn't exist")
	ErrSessionExpired    = Error("session expired")
)
