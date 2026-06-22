package engine

import "errors"

var (
	ErrConflict   = errors.New("conflict")
	ErrNoChanges  = errors.New("no changes")
	ErrNotFound   = errors.New("not found")
	ErrValidation = errors.New("validation failed")
)
