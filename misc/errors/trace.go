package errors

import "fmt"

func Trace(msg string, err error) error {
	return fmt.Errorf("%s failed: %v", msg, err)
}
