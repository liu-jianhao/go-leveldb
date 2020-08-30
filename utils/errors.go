package utils

import "errors"

var (
	ErrNotFound       = errors.New("Not Found")
	ErrDeletion       = errors.New("Type Deletetion")
	ErrTableFileMagic = errors.New("Wrong magic number")
)
