package lockservice

// Error provides constant error strings to the driver functions.
type Error string

func (e Error) Error() string { return string(e) }

// Constant errors.
// Rule of tthumb, all errors start with a small letter and end with no full stop.
const (
	ErrFileAcquired    = Error("file already acquired")
	ErrCantReleaseFile = Error("file cannot be released, wasn't locked before")
)
