package shrinkmyurl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var templates *template.Template

func init() {
	mustParseTemplates()
}

// Options for the Router service.
type RouterOptions struct {
	DevMode   bool
	Shortener *Shortener
}

// HTTP router for the service.
type Router struct {
	RouterOptions
}

// Create a new router with the given options.
func NewRouter(ops RouterOptions) *Router {
	if ops.Shortener == nil {
		panic(ErrShortenerRequired)
	}

	return &Router{RouterOptions: ops}
}

// Returns a new router with the routes defined.
func (rs *Router) Routes() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if !rs.DevMode {
		r.Use(redirectToHTTPS)
	}

	r.Get("/", rs.index)
	r.Post("/shorten", rs.shortenLink)
	r.Get("/{id}", rs.redirectLink)

	r.Get("/favicon.ico", rs.asset("favicon.ico"))

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", rs.apiHealthCheck)
		r.Post("/links", rs.apiShortenLink)
		r.Get("/links/{id}", rs.apiExpandLink)
	})

	return r
}

// Render and return the index page.
func (rs *Router) index(w http.ResponseWriter, r *http.Request) {
	rs.renderTemplate(w, "index.html", nil)
}

// Shorten the URL submitted via the form, render and return the shorten page.
func (rs *Router) shortenLink(w http.ResponseWriter, r *http.Request) {
	link := r.FormValue("url")

	record, err := rs.Shortener.Shorten(context.Background(), rs.requestURL(r), link)

	if err != nil {
		panic(err)
	}

	data := struct {
		Record Record
	}{
		Record: record,
	}

	rs.renderTemplate(w, "shorten.html", data)
}

// Visit the shortened URL and redirect to the expanded URL.
func (rs *Router) redirectLink(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	record, err := rs.Shortener.Expand(context.Background(), rs.requestURL(r), id)

	if err == ErrNil {
		http.NotFound(w, r)
	} else if err != nil {
		panic(err)
	} else {
		http.Redirect(w, r, record.ExpandedUrl, http.StatusFound)
	}
}

// Serve the asset file with caching enabled.
func (rs *Router) asset(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")

		http.ServeFile(w, r, filepath.Join("assets", name))
	}
}

// Check the health of the service.
func (rs *Router) apiHealthCheck(w http.ResponseWriter, r *http.Request) {
	err := rs.Shortener.Store.Ping(context.Background())

	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
}

// Shorten the URL submitted via JSON, return the shortened URL.
func (rs *Router) apiShortenLink(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	var payload Record

	err = json.Unmarshal(body, &payload)

	if err != nil {
		panic(err)
	}

	if !rs.Shortener.Validate(payload.ExpandedUrl) {
		handleError(w, ErrInvalidURL, http.StatusBadRequest)
		return
	}

	record, err := rs.Shortener.Shorten(context.Background(), rs.requestURL(r), payload.ExpandedUrl)

	if err != nil {
		panic(err)
	}

	writeJson(w, record, http.StatusCreated)
}

// Expand the shortened URL by ID, if it exists.
func (rs *Router) apiExpandLink(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	record, err := rs.Shortener.Expand(context.Background(), rs.requestURL(r), id)

	if err == ErrNil {
		http.NotFound(w, r)
	} else if err != nil {
		panic(err)
	} else {
		writeJson(w, record, http.StatusOK)
	}
}

// Get the request URL from the request, factoring in the scheme and host.
func (rs *Router) requestURL(r *http.Request) url.URL {
	if r.URL == nil {
		panic(ErrURLIsRequired)
	}

	host := *r.URL

	host.Host = r.Host
	host.Scheme = "http"

	if !rs.DevMode {
		host.Scheme = "https"
	}

	return host
}

// Render the template with the given name and data.
func (rs *Router) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if rs.DevMode {
		mustParseTemplates()
	}

	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		panic(err)
	}
}

// Write JSON to the response writer, with the given status code.
func writeJson(w http.ResponseWriter, data interface{}, status int) error {
	err := json.NewEncoder(w).Encode(data)

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
	} else {
		panic(err)
	}

	return err
}

// Handle an error by writing it to the response writer.
func handleError(w http.ResponseWriter, err error, status int) {
	http.Error(w, err.Error(), status)
}

// Parse templates from the HTML files.
func mustParseTemplates() {
	files, err := os.ReadDir("html")

	if err != nil {
		panic(err)
	}

	names := make([]string, 0, len(files))

	for _, file := range files {
		names = append(names, filepath.Join("html", file.Name()))
	}

	templates = template.Must(template.ParseFiles(names...))
}

// Redirect to HTTPS if the request is not secure.
func redirectToHTTPS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-Proto") != "https" {
			http.Redirect(w, r, fmt.Sprintf("https://%s%s", r.Host, r.RequestURI), http.StatusMovedPermanently)
			return
		}

		h.ServeHTTP(w, r)
	})
}
