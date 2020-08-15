package lockclient

// Error provides constant error strings to the driver functions.
type Error string

func (e Error) Error() string { return string(e) }

// Constant errors.
// Rule of thumb, all errors start with a small letter and end with no full stop.
const (
	ErrorObjectAlreadyPouncedOn = Error("the object already has a minimum of one pouncer")
)
