package errors

import (
	baseerr "errors"
	"fmt"
)

func New(msg string, vs ...interface{}) error {
	if len(vs) > 0 {
		msg = fmt.Sprintf(msg, vs...)
	}
	return baseerr.New(msg)
}
