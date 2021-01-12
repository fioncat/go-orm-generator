package errors

import "fmt"

// Trace adds "msg" information before the error
// message in order to trace the error.
func Trace(msg string, err error) error {
	return fmt.Errorf("%s failed: %v", msg, err)
}
