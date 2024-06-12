package shrinkmyurl_test

import (
	"testing"

	shrink "github.com/derek-schaefer/shrink-my-url"
	"github.com/stretchr/testify/assert"
)

func TestMust(t *testing.T) {
	assert.Equal(t, 1, shrink.Must(1, nil))

	assert.Panics(t, func() {
		shrink.Must(1, shrink.ErrNil)
	})
}
