package shrinkmyurl

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/sqids/sqids-go"
)

var (
	ErrInvalidURL   = errors.New("shortener: invalid URL")
	ErrDoesNotExist = errors.New("shortener: id does not exist")
	ErrMaxRetries   = errors.New("shortener: max retries exceeded")
)

// Represents a shortened URL record.
type Record struct {
	Id           string `json:"id"`
	Visits       int64  `json:"visits"`
	ExpandedUrl  string `json:"expanded_url"`
	ShortenedUrl string `json:"shortened_url"`
}

// Options for the Shortener service.
type ShortenerOptions struct {
	Store      Store
	Random     *rand.Rand
	MaxRetries uint
}

// Shortener is a service that shortens and expands URLs.
type Shortener struct {
	ShortenerOptions
}

// Create a new Shortener with the given options.
func NewShortener(ops ShortenerOptions) *Shortener {
	return &Shortener{ShortenerOptions: ops}
}

// Validate that the given link is a valid URL.
func (s *Shortener) Validate(link string) bool {
	_, err := url.ParseRequestURI(link)

	return err == nil
}

// Store a shortened URL using a random unique ID, and return the resulting record.
// If the ID already exists, retry until a unique ID is generated or the max retries is reached.
func (s *Shortener) Shorten(ctx context.Context, host url.URL, link string) (Record, error) {
	if !s.Validate(link) {
		return Record{}, ErrInvalidURL
	}

	var retries uint

	for {
		if retries > s.MaxRetries {
			return Record{}, ErrMaxRetries
		}

		id, err := s.generateId()

		if err != nil {
			return Record{}, err
		}

		ok, err := s.Store.AddLink(ctx, id, link)

		if err != nil && err != ErrExists {
			return Record{}, err
		}

		if ok {
			return Record{Id: id, ExpandedUrl: link, ShortenedUrl: shortenedUrl(host, id)}, nil
		}

		retries++
	}
}

// Expand the shortened URL by ID, if it exists, and increment the visit count.
func (s *Shortener) Expand(ctx context.Context, host url.URL, id string) (Record, error) {
	link, visits, err := s.Store.ExpandLink(ctx, id)

	if err != nil {
		return Record{}, err
	}

	record := Record{
		Id:           id,
		ExpandedUrl:  link,
		ShortenedUrl: shortenedUrl(host, id),
		Visits:       visits,
	}

	return record, nil
}

// Generate a unique ID for the shortened URL.
func (s *Shortener) generateId() (string, error) {
	ids, err := sqids.New()

	if err != nil {
		return "", err
	}

	return ids.Encode([]uint64{s.Random.Uint64()})
}

// Return the shortened URL for the given ID.
func shortenedUrl(url url.URL, id string) string {
	url.Path = fmt.Sprintf("/%s", id)

	return url.String()
}
