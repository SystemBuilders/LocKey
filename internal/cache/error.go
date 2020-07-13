package cache

// Error provides constant error strings to the driver functions.
type Error string

func (e Error) Error() string { return string(e) }

// Constant errors.
// Rule of thumb, all errors start with a small letter and end with no full stop.
const (
	ErrElementDoesntExist   = Error("element doesn't exist in the cache")
	ErrElementAlreadyExists = Error("element already exists in the cache")
)
