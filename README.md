# Shrink My URL

https://shrink.derekschaefer.com

Yet another URL shortening service, implemented in Go using Redis as a database.

The HTTP server provides a simple UI and JSON API. By default, the server redirects to HTTPS which is disabled in dev mode. The UI uses [Tailwind](https://tailwindcss.com/) and [htmx](https://htmx.org/).

The shortened IDs are randomly generated using [sqids](https://sqids.org/). They are relatively short and have a low collision chance, although collisions are handled.

Visits to shortened URLs are counted. By default, shortened links have a TTL of 24 hours to keep the datastore tidy.

## Routes

- `GET /`: Renders the home page.
- `POST /shorten`: Shortens the submitted URL and renders a page fragment. Expects a form.
- `GET /{id}`: Expands and redirects to the shortened URL, if it exists.
- `GET /api/health`: A health check endpoint that tests the Redis connection.
- `POST /api/links`: Shortens the submitted URL. Expects and returns JSON.
- `GET /api/links/{id}`: Expands and returns the shortened URL, if it exists. Returns JSON.

## Development

A [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers) configuration is included and recommended. Alternatively, you can use Docker Compose directly.

Run the server:

```
$ go run cmd/main.go -dev
```

The server supports various options:

```
$ go run cmd/main.go -h
```

Run the tests:

```
$ go test ./...
```

## Testing

The Github repo has Actions enabled and configured. The workflows can be found [here](https://github.com/derek-schaefer/shrink-my-url/tree/main/.github/workflows).

## Deployment

The service is deployed to [Heroku](https://shrink-my-url-be1f27a69f91.herokuapp.com/) from Github Actions. The workflow is [here](https://github.com/derek-schaefer/shrink-my-url/blob/main/.github/workflows/deploy.yml).

## Dependencies

- Go
- Redis
- Docker
- Heroku

## License

MIT
