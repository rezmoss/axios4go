# Threat model: axios4go

## What this project does

axios4go is a Go HTTP client library modelled on axios (`github.com/rezmoss/axios4go`). It wraps `net/http` with
package-level helpers (`Get`, `Post`, `Request`, ... and `*Async` promise variants), a `Client` with a base URL,
request/response interceptors, proxy support, basic auth, upload/download progress callbacks, an in-memory response
cache (`cache.go`), and a logger that masks sensitive headers (`logger.go`). There is no server component and no CLI.

## Where untrusted input enters

Treat as attacker-controlled:

- **The remote server and anything on the path to it**: status line, headers (`Location`, `Cache-Control`,
  `Content-Encoding`, `Set-Cookie`), body size and content, compressed bodies, slow or never-ending responses, and
  redirect chains (cross-host, scheme downgrade, loops).
- **The relative URL and query params** an application passes to a `Client` with a base URL, since applications
  commonly build them from their own users' input. The base URL itself and the rest of `RequestOptions` are trusted
  developer configuration.
- **Values that end up in logs**: URLs, headers and bodies logged by `DefaultLogger`.

## Security properties the library promises

- A relative URL joined onto a base URL must stay under the base path (`joinBaseURL`, `hasDotDotSegment`), including
  percent-encoded and mixed-encoding `..` tricks, and must not switch host or scheme.
- `Authorization`, `Cookie`, `Proxy-Authorization` and other non-allowlisted headers are stripped when a redirect
  leaves the original host (or its subdomains) or downgrades https to http (`stripCredentialsOnRedirect`).
  `MaxRedirects` is enforced.
- `MaxContentLength` bounds the response body after decompression; `MaxBodyLength` bounds the request body.
  Negative limits are rejected.
- Proxy host/port/protocol are validated (`buildProxyURL`); proxy credentials go only to the proxy.
- The cache never serves one request's response for a different request (key covers method, full URL, a digest of all
  request headers and the body hash), does not store `Cache-Control: no-store` or non-2xx responses, returns copies so callers
  cannot mutate cached entries, and evicts within `MaxEntries`.
- The logger masks sensitive headers and neutralises CR/LF so remote data cannot forge log
  lines.
- Panics in promise callbacks (`Then`, `Catch`, `Finally`) are recovered rather than crashing
  the process; concurrent use of a `Client` is race-free.

## Severity

- **Critical**: credentials (auth header, cookies, proxy auth) sent to a host other than the one the developer
  targeted; a cached response returned for a different user's or request's key.
- **High**: base-URL escape to another path, host or scheme via a relative URL; unbounded memory from a hostile
  server despite `MaxContentLength` (e.g. decompression bomb); a data race or a panic reachable from server input
  that takes down the process.
- **Medium**: log injection or unmasked secrets in logs; redirect-limit bypass; goroutine or connection leaks per
  request; cache growth beyond configured limits.
- **Low**: issues that need a non-default, explicitly unsafe option or a hostile developer configuration.

## Scope

In scope: `client.go`, `cache.go`, `logger.go`. Out of scope: `examples/`, `*_test.go`, `dist/`, CI config.
Flaws wholly inside `net/http` or the Go standard library belong upstream; report them only if axios4go's usage
makes them reachable and show the path. There are no third-party module dependencies.

## How to exercise it

The image has only loopback networking. Everything is prebuilt in `/src`; `GOPROXY=off`, `GOTOOLCHAIN=local`.

- Tests: `go test -race ./...`. `security_test.go` holds the existing security regression tests; new reproducers fit
  best as Go tests using `net/http/httptest` servers (redirect chains, gzip bombs, oversized bodies).
- Examples are built under `.tmp/bin/`; most target public URLs and will not work offline.

## How to report

Give a Go test reproducer that runs in this image against an `httptest` server, state any non-default
`RequestOptions`/`CacheConfig` used, and propose a minimal patch against `main` with a regression test. One root
cause, one report.
