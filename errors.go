package goutil

import (
	"errors"
	"fmt"
)

func NewError(v ...interface{}) error {
	return errors.New(fmt.Sprint(v...))
}

func NewErrorf(format string, v ...interface{}) error {
	return fmt.Errorf(format, v...)
}

func NewError2(err error, v ...interface{}) error {
	return fmt.Errorf("%v: %s", err, fmt.Sprint(v...))
}

func NewError2f(err error, format string, v ...interface{}) error {
	return fmt.Errorf("%v: %s", err, fmt.Sprintf(format, v...))
}
