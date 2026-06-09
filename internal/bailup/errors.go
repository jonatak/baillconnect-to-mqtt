package bailup

import "fmt"

type BailupError struct {
	message string
	inner   error
}

func NewBailupError(message string, inner error) *BailupError {
	return &BailupError{
		message: message,
		inner:   inner,
	}
}

func (e *BailupError) Error() string {
	if e.inner != nil {
		return fmt.Sprintf("%s: %v", e.message, e.inner)
	}
	return e.message
}

func (e *BailupError) Unwrap() error {
	return e.inner
}
