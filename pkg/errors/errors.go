package errors

import (
	"errors"
	"fmt"
)

var (
	ErrorIs = errors.Is

	ErrPriceNotFound = fmt.Errorf("price not found")
	ErrInternal      = fmt.Errorf("internal error")
)
