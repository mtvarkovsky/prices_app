package errors

import (
	"errors"
	"fmt"
)

var (
	ErrorIs = errors.Is

	PriceNotFound = fmt.Errorf("price not found")
	InternalError = fmt.Errorf("internal error")
)
