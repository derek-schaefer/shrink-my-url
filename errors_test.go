package shrinkmyurl_test

import (
	"testing"

	shrink "github.com/derek-schaefer/shrink-my-url"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	assert.Equal(t, 1, shrink.Must(1, nil))

	assert.Panics(t, func() {
		shrink.Must(1, shrink.ErrNil)
	})
}

func TestStoreError(t *testing.T) {
	assert.Equal(t, shrink.ErrNil, shrink.NormalizeError(redis.Nil))
	assert.Equal(t, shrink.ErrExists, shrink.NormalizeError(shrink.ErrExists))
}
