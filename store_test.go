package shrinkmyurl_test

import (
	"context"
	"testing"
	"time"

	shrink "github.com/derek-schaefer/shrink-my-url"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type testRedisStore struct {
	*shrink.RedisStore
}

func newTestRedisStore() *testRedisStore {
	store := shrink.Must(shrink.NewRedisStore(shrink.RedisStoreOptions{
		Expiration: 1 * time.Minute,
	}))

	return &testRedisStore{
		RedisStore: store,
	}
}

func TestMemoryStore(t *testing.T) {
	store := shrink.Store(shrink.NewMemoryStore())

	assert.Nil(t, store.Close())
	assert.Nil(t, store.Ping(context.Background()))

	ok, err := store.AddLink(context.Background(), "id", "url")

	assert.True(t, ok)
	assert.Nil(t, err)

	link, visits, err := store.ExpandLink(context.Background(), "id")

	assert.Equal(t, "url", link)
	assert.Equal(t, int64(1), visits)
	assert.Nil(t, err)
}

func TestRedisStoreClose(t *testing.T) {
	store := newTestRedisStore()

	assert.Nil(t, store.Close())
}

func TestRedisStorePing(t *testing.T) {
	store := newTestRedisStore()

	defer store.Close()

	assert.Nil(t, store.Ping(context.Background()))
}

func TestRedisStoreAddLink(t *testing.T) {
	store := newTestRedisStore()

	defer store.Close()
	defer store.DeleteLink(context.Background(), "id")

	ok, err := store.AddLink(context.Background(), "id", "url")

	assert.True(t, ok)
	assert.Nil(t, err)

	ok, err = store.AddLink(context.Background(), "id", "url")

	assert.False(t, ok)
	assert.Equal(t, shrink.ErrExists, err)
}

func TestRedisStoreExpandLink(t *testing.T) {
	store := newTestRedisStore()

	defer store.Close()
	defer store.DeleteLink(context.Background(), "id")

	ok, err := store.AddLink(context.Background(), "id", "url")

	assert.True(t, ok)
	assert.Nil(t, err)

	link, visits, err := store.ExpandLink(context.Background(), "id")

	assert.Equal(t, "url", link)
	assert.Equal(t, int64(1), visits)
	assert.Nil(t, err)

	link, visits, err = store.ExpandLink(context.Background(), "id")

	assert.Equal(t, "url", link)
	assert.Equal(t, int64(2), visits)
	assert.Nil(t, err)
}

func TestRedisStoreDeleteLink(t *testing.T) {
	store := newTestRedisStore()

	defer store.Close()
	defer store.DeleteLink(context.Background(), "id")

	ok, err := store.AddLink(context.Background(), "id", "url")

	assert.True(t, ok)
	assert.Nil(t, err)

	err = store.DeleteLink(context.Background(), "id")

	assert.Nil(t, err)

	link, visits, err := store.ExpandLink(context.Background(), "id")

	assert.Equal(t, "", link)
	assert.Equal(t, int64(0), visits)
	assert.NotNil(t, err)
}

func TestStoreError(t *testing.T) {
	assert.Equal(t, shrink.ErrNil, shrink.StoreError(redis.Nil))
	assert.Equal(t, shrink.ErrExists, shrink.StoreError(shrink.ErrExists))
}
