package lockservice

// Error provides constant error strings to the driver functions.
type Error string

func (e Error) Error() string { return string(e) }

// Constant errors.
// Rule of thumb, all errors start with a small letter and end with no full stop.
const (
	ErrFileAcquired        = Error("file already acquired")
	ErrCantReleaseFile     = Error("file cannot be released, wasn't locked before")
	ErrUnauthorizedAccess  = Error("file cannot be released, unauthorized access")
	ErrCheckAcquireFailure = Error("file is not acquired")
	ErrFileUnlocked        = Error("file doesn't have a lock")
)
