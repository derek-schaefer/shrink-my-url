package shrinkmyurl

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNil    = errors.New("store: key not found")
	ErrExists = errors.New("store: key already exists")
)

// A service for shortening and expanding URLs.
type Store interface {
	Close() error
	Ping(ctx context.Context) error
	AddLink(ctx context.Context, id, url string) (bool, error)
	ExpandLink(ctx context.Context, id string) (string, int64, error)
	DeleteLink(ctx context.Context, id string) error
}

// A simple in-memory store implementation.
type MemoryStore struct {
	links  map[string]string
	visits map[string]int64
}

// Create a new memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		links:  make(map[string]string),
		visits: make(map[string]int64),
	}
}

// Close the memory store.
func (s *MemoryStore) Close() error {
	return nil
}

// Ping the memory store.
func (s *MemoryStore) Ping(ctx context.Context) error {
	return nil
}

// Add a link to the memory store.
func (s *MemoryStore) AddLink(ctx context.Context, id, url string) (bool, error) {
	if _, ok := s.links[id]; ok {
		return false, ErrExists
	}

	s.links[id] = url
	s.visits[id] = 0

	return true, nil
}

// Expand a shortened link from the memory store with the number of visits, incrementing the visit count.
func (s *MemoryStore) ExpandLink(ctx context.Context, id string) (string, int64, error) {
	if url, ok := s.links[id]; ok {
		s.visits[id]++
		return url, s.visits[id], nil
	}

	return "", 0, ErrNil
}

// Delete a link and its visit count from the memory store.
func (s *MemoryStore) DeleteLink(ctx context.Context, id string) error {
	delete(s.links, id)
	delete(s.visits, id)

	return nil
}

// Options for the Redis store.
type RedisStoreOptions struct {
	redis.Options

	Expiration time.Duration
}

// A Store implementation that uses Redis.
type RedisStore struct {
	RedisStoreOptions

	client *redis.Client
}

// Create a new Redis store with the given options.
func NewRedisStore(ops RedisStoreOptions) (*RedisStore, error) {
	client := redis.NewClient(&ops.Options)

	return &RedisStore{RedisStoreOptions: ops, client: client}, nil
}

// Ping the Redis store.
func (s *RedisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Close the Redis store.
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Add a link to the store with the configured expiration.
// Returns true if the link was successfully added, or false if it was not.
func (s *RedisStore) AddLink(ctx context.Context, id, url string) (bool, error) {
	cmds, err := s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.SetNX(ctx, id, url, s.Expiration)
		p.SetNX(ctx, visitId(id), 0, s.Expiration)

		return nil
	})

	if err != nil {
		return false, StoreError(err)
	}

	var exists bool

	for _, c := range cmds {
		if !c.(*redis.BoolCmd).Val() {
			exists = true
		}
	}

	if exists {
		return false, ErrExists
	}

	return true, nil
}

// Expand a shortened link from the store with the number of visits, incrementing the visit count.
func (s *RedisStore) ExpandLink(ctx context.Context, id string) (string, int64, error) {
	vid := visitId(id)

	cmds, err := s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.GetEx(ctx, id, s.Expiration)
		p.Incr(ctx, vid)
		p.Expire(ctx, vid, s.Expiration)

		return nil
	})

	if err != nil {
		return "", 0, StoreError(err)
	}

	link := cmds[0].(*redis.StringCmd).Val()
	count := cmds[1].(*redis.IntCmd).Val()

	return link, count, nil
}

// Delete a link and its visit count from the store.
func (s *RedisStore) DeleteLink(ctx context.Context, id string) error {
	_, err := s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		p.Del(ctx, id)
		p.Del(ctx, visitId(id))

		return nil
	})

	return StoreError(err)
}

// Get the visits count ID for the given link ID.
func visitId(id string) string {
	return fmt.Sprintf("%s:visits", id)
}

// Converts errors to normalized values.
func StoreError(err error) error {
	if err == redis.Nil {
		return ErrNil
	}

	return err
}
