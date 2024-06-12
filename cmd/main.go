package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	shrink "github.com/derek-schaefer/shrink-my-url"
	"github.com/redis/go-redis/v9"
)

func main() {
	httpAddr := flag.String("httpAddr", ":8080", "address to listen on")
	redisAddr := flag.String("redisAddr", "redis://localhost:6379/0", "address of the redis server")
	expiration := flag.Duration("expiration", 24*time.Hour, "expiration time for shortened URLs")
	maxRetries := flag.Uint("maxRetries", 5, "maximum number of retries when generating a short URL")
	devMode := flag.Bool("dev", false, "enable development mode")

	flag.Parse()

	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	redisOptions, err := redis.ParseURL(*redisAddr)

	if err != nil {
		log.Fatal(err)
	}

	store, err := shrink.NewRedisStore(shrink.RedisStoreOptions{
		Options:    *redisOptions,
		Expiration: *expiration,
	})

	if err != nil {
		log.Fatal(err)
	}

	defer store.Close()

	err = store.Ping(context.Background())

	if err != nil {
		log.Fatal(err)
	}

	shortener := shrink.NewShortener(shrink.ShortenerOptions{
		Store:      store,
		Random:     random,
		MaxRetries: *maxRetries,
	})

	router := shrink.NewRouter(shrink.RouterOptions{
		DevMode:   *devMode,
		Shortener: shortener,
	})

	log.Fatal(http.ListenAndServe(*httpAddr, router.Routes()))
}
