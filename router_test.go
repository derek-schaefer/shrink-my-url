package shrinkmyurl_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	shrink "github.com/derek-schaefer/shrink-my-url"
	"github.com/stretchr/testify/assert"
)

var (
	localURL = url.URL{
		Scheme: "http",
		Host:   "example.com",
	}
)

type testRouter struct {
	*shrink.Router

	store     shrink.Store
	shortener *shrink.Shortener
}

func newTestRouter() *testRouter {
	store := shrink.NewMemoryStore()
	random := rand.New(rand.NewSource(0))

	shortener := shrink.NewShortener(shrink.ShortenerOptions{
		Store:  store,
		Random: random,
	})

	router := shrink.NewRouter(shrink.RouterOptions{
		DevMode:   true,
		Shortener: shortener,
	})

	return &testRouter{
		Router:    router,
		store:     store,
		shortener: shortener,
	}
}

func TestNewRouter(t *testing.T) {
	assert.NotNil(t, shrink.NewRouter(shrink.RouterOptions{
		Shortener: shrink.NewShortener(shrink.ShortenerOptions{
			Store: shrink.NewMemoryStore(),
		}),
	}))

	assert.Panics(t, func() {
		shrink.NewRouter(shrink.RouterOptions{})
	})
}

func TestRouterRoutes(t *testing.T) {
	router := newTestRouter()

	assert.NotNil(t, router.Routes())
}

func TestRouterIndex(t *testing.T) {
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := recordRequest(router, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Shrink My URL")
}

func TestRouterShorten(t *testing.T) {
	router := newTestRouter()

	form := url.Values{
		"url": []string{"http://example.com"},
	}

	request := postForm("/shorten", form)
	recorder := recordRequest(router, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Here's your link")
	assert.Contains(t, recorder.Body.String(), form.Get("url"))
}

func TestRouterRedirect(t *testing.T) {
	router := newTestRouter()

	record := shrink.Must(router.shortener.Shorten(context.Background(), localURL, "http://example.com"))

	request := httptest.NewRequest(http.MethodGet, "/"+record.Id, nil)
	recorder := recordRequest(router, request)

	assert.Equal(t, http.StatusFound, recorder.Code)
	assert.Equal(t, record.ExpandedUrl, recorder.Header().Get("Location"))
}

func TestRouterApiHealth(t *testing.T) {
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := recordRequest(router, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestRouterApiShorten(t *testing.T) {
	router := newTestRouter()

	payload := shrink.Record{
		ExpandedUrl: "http://example.com",
	}

	request := postJSON("/api/links", payload)
	recorder := recordRequest(router, request)

	var received shrink.Record
	unmarshalJSON(recorder.Body.Bytes(), &received)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, payload.ExpandedUrl, received.ExpandedUrl)
}

func TestRouterApiExpand(t *testing.T) {
	router := newTestRouter()

	record := shrink.Must(router.shortener.Shorten(context.Background(), localURL, "http://example.com"))

	request := httptest.NewRequest(http.MethodGet, "/api/links/"+record.Id, nil)
	recorder := recordRequest(router, request)

	var received shrink.Record
	unmarshalJSON(recorder.Body.Bytes(), &received)
	record.Visits = 1

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, record, received)
}

func recordRequest(router *testRouter, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.Routes().ServeHTTP(recorder, request)
	return recorder
}

func postForm(url string, form url.Values) *http.Request {
	request := httptest.NewRequest(http.MethodPost, url, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return request
}

func postJSON(url string, payload interface{}) *http.Request {
	request := httptest.NewRequest(http.MethodPost, url, marshalJSON(payload))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func marshalJSON(v interface{}) *bytes.Buffer {
	data, err := json.Marshal(v)

	if err != nil {
		panic(err)
	}

	return bytes.NewBuffer(data)
}

func unmarshalJSON(data []byte, v interface{}) {
	if err := json.Unmarshal(data, v); err != nil {
		panic(err)
	}
}
