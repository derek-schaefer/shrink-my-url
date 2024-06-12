package shrinkmyurl_test

import (
	"context"
	"math/rand"
	"net/url"
	"testing"

	shrink "github.com/derek-schaefer/shrink-my-url"
	"github.com/stretchr/testify/assert"
)

type testShortener struct {
	*shrink.Shortener

	store shrink.Store
}

func (t *testShortener) resetRandom() {
	t.Random = rand.New(rand.NewSource(0))
}

func newTestShortener() *testShortener {
	store := shrink.NewMemoryStore()

	shortener := &testShortener{
		Shortener: shrink.NewShortener(shrink.ShortenerOptions{
			Store: store,
		}),
		store: store,
	}

	shortener.resetRandom()

	return shortener
}

func TestNewShortener(t *testing.T) {
	shortener := newTestShortener()

	assert.NotNil(t, shortener)
	assert.Equal(t, shortener.store, shortener.Store)
}

func TestShortenerValidate(t *testing.T) {
	shortener := newTestShortener()

	assert.False(t, shortener.Validate("asdf"))
	assert.True(t, shortener.Validate("http://asdf.com"))
}

func TestShortenerShorten(t *testing.T) {
	shortener := newTestShortener()

	url := url.URL{
		Scheme: "http",
		Host:   "example.com",
	}

	record := shrink.Must(shortener.Shorten(context.Background(), url, "http://asdf.com"))

	defer shortener.store.DeleteLink(context.Background(), record.Id)

	assert.Equal(t, "Aq3vRfb3kuc2", record.Id)
	assert.Equal(t, "http://example.com/Aq3vRfb3kuc2", record.ShortenedUrl)
	assert.Equal(t, "http://asdf.com", record.ExpandedUrl)

	shortener.resetRandom()

	record, err := shortener.Shorten(context.Background(), url, "http://asdf.com")

	assert.Equal(t, shrink.ErrMaxRetries, err)
}

func TestShortenerExpand(t *testing.T) {
	shortener := newTestShortener()

	url := url.URL{
		Scheme: "http",
		Host:   "example.com",
	}

	original := shrink.Must(shortener.Shorten(context.Background(), url, "http://asdf.com"))

	defer shortener.store.DeleteLink(context.Background(), original.Id)

	record, err := shortener.Expand(context.Background(), url, original.Id)

	assert.Nil(t, err)
	assert.Equal(t, original.Id, record.Id)
	assert.Equal(t, "http://example.com/"+original.Id, record.ShortenedUrl)
	assert.Equal(t, "http://asdf.com", record.ExpandedUrl)
}
