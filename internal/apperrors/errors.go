package apperrors

import (
	"errors"
)

var (
	ErrShutdown = errors.New("shutdown error")

	ErrTestDataDoesNotExist = errors.New("test data does not exist")
)
