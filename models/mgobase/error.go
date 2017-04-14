package mgobase

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
)

const (
	// ErrInvalidID represents the id is not a valid object id.
	ErrInvalidID ModelError = iota + 1
	// ErrNotFound represents the record retrieved not found.
	ErrNotFound
	// ErrDuplicateKey represents the document to be inserted or updated into db has conflicting value on a field.
	ErrDuplicateKey
	// ErrNotConnected represents can't not connect to db.
	ErrNotConnected
)

// ModelError is the mgobase package level error type.
type ModelError int

func (e ModelError) Error() string {
	switch e {
	case ErrInvalidID:
		return "invalid object id"
	case ErrDuplicateKey:
		return "duplicate key"
	case ErrNotFound:
		return "not found"
	case ErrNotConnected:
		return "db is not connected"
	default:
		return fmt.Sprintf("undefined model error, number: %d", int(e))
	}
}

func parseMgoError(err error) error {
	if err == mgo.ErrNotFound {
		return ErrNotFound
	}

	if mgo.IsDup(err) {
		return ErrDuplicateKey
	}

	return err
}
