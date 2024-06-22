package shrinkmyurl

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	ErrDoesNotExist      = errors.New("shortener: id does not exist")
	ErrExists            = errors.New("store: key already exists")
	ErrInvalidURL        = errors.New("shortener: invalid URL")
	ErrMaxRetries        = errors.New("shortener: max retries exceeded")
	ErrNil               = errors.New("store: key not found")
	ErrShortenerRequired = errors.New("router: shortener is required")
	ErrURLIsRequired     = errors.New("router: URL is required")
)

// Panics if the given error is not nil.
func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

// Converts errors to normalized values.
func NormalizeError(err error) error {
	if err == redis.Nil {
		return ErrNil
	}

	return err
}
