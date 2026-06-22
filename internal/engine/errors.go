package engine

import "errors"

var (
	ErrConflict   = errors.New("conflict")
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation failed")
)
