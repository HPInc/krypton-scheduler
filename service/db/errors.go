package db

import "errors"

var (
	ErrTokenFileNotFound   = errors.New("role token file was not found")
	ErrInvalidDatabaseType = errors.New("invalid keyspace provider specified in configuration")
	ErrDuplicateEntry      = errors.New("a duplicate entry was found in the database")
	ErrNotFound            = errors.New("the requested entry was not found in the database")
	ErrNotAllowed          = errors.New("the requested operation is not allowed")
	ErrInvalidRequest      = errors.New("the request contained one or more invalid parameters")
	ErrInternalError       = errors.New("an internal error occured while performing the database operation")
)
