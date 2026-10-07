package axios4go

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	Logger      Logger
	CacheConfig *CacheConfig

	// mu guards BaseURL against concurrent SetBaseURL calls.
	mu sync.RWMutex
	// proxyTransports caches one *http.Transport per proxy configuration so
	// proxied requests reuse connections instead of leaking a pool per call.
	proxyTransports sync.Map
}

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

type Promise struct {
	response *Response
	err      error
	then     func(*Response)
	catch    func(error)
	finally  func()
	done     chan struct{}
	settled  bool
	mu       sync.Mutex
}

type RequestInterceptors []func(*http.Request) error
type ResponseInterceptors []func(*http.Response) error
type InterceptorOptions struct {
	RequestInterceptors  RequestInterceptors
	ResponseInterceptors ResponseInterceptors
}

type RequestOptions struct {
	Method               string
	URL                  string
	BaseURL              string
	Params               map[string]string
	Body                 interface{}
	Headers              map[string]string
	Timeout              int
	Auth                 *Auth
	ResponseType         string
	ResponseEncoding     string
	MaxRedirects         int
	MaxContentLength     int64
	MaxBodyLength        int64
	Decompress           bool
	DisableDecompression bool
	ValidateStatus       func(int) bool
	InterceptorOptions   InterceptorOptions
	Proxy                *Proxy
	OnUploadProgress     func(bytesRead, totalBytes int64)
	OnDownloadProgress   func(bytesRead, totalBytes int64)
	LogLevel             LogLevel
	Cache                *RequestCacheOptions
}

type Proxy struct {
	Protocol string
	Host     string
	Port     int
	Auth     *Auth
}

type Auth struct {
	Username string
	Password string
}

type ProgressReader struct {
	reader     io.Reader
	total      int64
	read       int64
	onProgress func(bytesRead, totalBytes int64)
}

type ProgressWriter struct {
	writer     io.Writer
	total      int64
	written    int64
	onProgress func(bytesWritten, totalBytes int64)
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.read += int64(n)
	if pr.onProgress != nil {
		pr.onProgress(pr.read, pr.total)
	}
	return n, err
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	pw.written += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.written, pw.total)
	}
	return n, err
}

var defaultClient = &Client{HTTPClient: &http.Client{}, Logger: NewLogger(LevelNone)}

func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// safeCall runs fn and converts a panic into an error.
func safeCall(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("promise callback panicked: %v", r)
		}
	}()
	fn()
	return nil
}

func (p *Promise) Then(fn func(*Response)) *Promise {
	p.mu.Lock()
	if p.settled && p.err == nil {
		response := p.response
		p.mu.Unlock()
		if perr := safeCall(func() { fn(response) }); perr != nil {
			p.mu.Lock()
			p.err = perr
			p.response = nil
			p.mu.Unlock()
		}
	} else {
		if !p.settled {
			p.then = fn
		}
		p.mu.Unlock()
	}
	return p
}

func (p *Promise) Catch(fn func(error)) *Promise {
	p.mu.Lock()
	if p.settled && p.err != nil {
		err := p.err
		p.mu.Unlock()
		fn(err)
	} else {
		if !p.settled {
			p.catch = fn
		}
		p.mu.Unlock()
	}
	return p
}

func (p *Promise) Finally(fn func()) {
	p.mu.Lock()
	if p.settled {
		p.mu.Unlock()
		fn()
	} else {
		p.finally = fn
		p.mu.Unlock()
	}

	<-p.done
}

func NewPromise() *Promise {
	return &Promise{
		done: make(chan struct{}),
	}
}

func (p *Promise) resolve(resp *Response, err error) {
	p.mu.Lock()
	if p.settled {
		p.mu.Unlock()
		return
	}

	p.response = resp
	p.err = err
	p.settled = true
	thenFn := p.then
	catchFn := p.catch
	finallyFn := p.finally
	p.mu.Unlock()

	// Always release Finally waiters, even if a callback panics.
	defer close(p.done)

	if thenFn != nil && err == nil {
		if perr := safeCall(func() { thenFn(resp) }); perr != nil {
			// A panicking Then callback rejects the promise, mirroring a
			// throw inside a JavaScript then handler.
			err = perr
			p.mu.Lock()
			p.err = perr
			p.response = nil
			p.mu.Unlock()
		}
	}
	if catchFn != nil && err != nil {
		catchFn(err)
	}
	if finallyFn != nil {
		finallyFn()
	}
}

func Get(urlStr string, options ...*RequestOptions) (*Response, error) {
	return Request("GET", urlStr, options...)
}

func GetAsync(urlStr string, options ...*RequestOptions) *Promise {
	promise := NewPromise()

	go func() {
		resp, err := Request("GET", urlStr, options...)
		promise.resolve(resp, err)
	}()

	return promise
}

func Post(urlStr string, body interface{}, options ...*RequestOptions) (*Response, error) {
	mergedOptions := mergeBodyIntoOptions(body, options)
	return Request("POST", urlStr, mergedOptions)
}

func PostAsync(urlStr string, body interface{}, options ...*RequestOptions) *Promise {
	mergedOptions := mergeBodyIntoOptions(body, options)
	promise := NewPromise()

	go func() {
		resp, err := Request("POST", urlStr, mergedOptions)
		promise.resolve(resp, err)
	}()

	return promise
}

func mergeBodyIntoOptions(body interface{}, options []*RequestOptions) *RequestOptions {
	mergedOption := &RequestOptions{
		Body: body,
	}

	if len(options) > 0 {
		*mergedOption = *options[0]
		mergedOption.Body = body
	}

	return mergedOption
}

func Put(urlStr string, body interface{}, options ...*RequestOptions) (*Response, error) {
	mergedOptions := mergeBodyIntoOptions(body, options)
	return Request("PUT", urlStr, mergedOptions)
}

func PutAsync(urlStr string, body interface{}, options ...*RequestOptions) *Promise {
	mergedOptions := mergeBodyIntoOptions(body, options)
	promise := NewPromise()

	go func() {
		resp, err := Request("PUT", urlStr, mergedOptions)
		promise.resolve(resp, err)
	}()

	return promise
}

func Delete(urlStr string, options ...*RequestOptions) (*Response, error) {
	return Request("DELETE", urlStr, options...)
}

func DeleteAsync(urlStr string, options ...*RequestOptions) *Promise {
	promise := NewPromise()
	go func() {
		resp, err := Request("DELETE", urlStr, options...)
		promise.resolve(resp, err)
	}()
	return promise
}

func Head(urlStr string, options ...*RequestOptions) (*Response, error) {
	return Request("HEAD", urlStr, options...)
}

func HeadAsync(urlStr string, options ...*RequestOptions) *Promise {
	promise := NewPromise()
	go func() {
		resp, err := Request("HEAD", urlStr, options...)
		promise.resolve(resp, err)
	}()
	return promise
}

func Options(urlStr string, options ...*RequestOptions) (*Response, error) {
	return Request("OPTIONS", urlStr, options...)
}

func OptionsAsync(urlStr string, options ...*RequestOptions) *Promise {
	promise := NewPromise()
	go func() {
		resp, err := Request("OPTIONS", urlStr, options...)
		promise.resolve(resp, err)
	}()
	return promise
}

func Patch(urlStr string, body interface{}, options ...*RequestOptions) (*Response, error) {
	mergedOptions := mergeBodyIntoOptions(body, options)
	return Request("PATCH", urlStr, mergedOptions)
}

func PatchAsync(urlStr string, body interface{}, options ...*RequestOptions) *Promise {
	mergedOptions := mergeBodyIntoOptions(body, options)
	promise := NewPromise()

	go func() {
		resp, err := Request("PATCH", urlStr, mergedOptions)
		promise.resolve(resp, err)
	}()

	return promise
}

func Request(method, urlStr string, options ...*RequestOptions) (*Response, error) {
	reqOptions := &RequestOptions{
		Method:           "GET",
		URL:              urlStr,
		Timeout:          1000,
		ResponseType:     "json",
		ResponseEncoding: "utf8",
		MaxContentLength: 2000,
		MaxBodyLength:    2000,
		MaxRedirects:     21,
		Decompress:       true,
		ValidateStatus:   nil,
	}

	if len(options) > 0 && options[0] != nil {
		mergeOptions(reqOptions, options[0])
	}

	if method != "" {
		reqOptions.Method = method
	}

	return defaultClient.Request(reqOptions)
}

func RequestAsync(method, urlStr string, options ...*RequestOptions) *Promise {
	promise := NewPromise()
	go func() {
		resp, err := Request(method, urlStr, options...)
		promise.resolve(resp, err)
	}()
	return promise
}

// safeRedirectHeaders are forwarded to redirect targets on another host or
// over a downgraded scheme. Everything else is treated as potentially
// sensitive and dropped, since net/http only strips a handful of well-known
// credential headers on its own.
var safeRedirectHeaders = map[string]bool{
	"Accept":          true,
	"Accept-Language": true,
	"Accept-Encoding": true,
	"Content-Type":    true,
	"Content-Length":  true,
	"User-Agent":      true,
	"Cache-Control":   true,
	"Referer":         true,
}

var validMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"PATCH":   true,
	"HEAD":    true,
	"OPTIONS": true,
}

var validProxyProtocols = map[string]bool{
	"http":    true,
	"https":   true,
	"socks5":  true,
	"socks5h": true,
}

func (c *Client) baseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.BaseURL
}

// cloneRequestOptions returns a copy of options whose maps are not shared with
// the caller, so Request never mutates caller-owned state.
func cloneRequestOptions(options *RequestOptions) *RequestOptions {
	opts := *options
	if options.Headers != nil {
		opts.Headers = make(map[string]string, len(options.Headers))
		for k, v := range options.Headers {
			opts.Headers[k] = v
		}
	}
	if options.Params != nil {
		opts.Params = make(map[string]string, len(options.Params))
		for k, v := range options.Params {
			opts.Params[k] = v
		}
	}
	return &opts
}

func applyDefaults(options *RequestOptions) error {
	switch {
	case options.Timeout < 0:
		return fmt.Errorf("invalid Timeout %d: must not be negative", options.Timeout)
	case options.MaxContentLength < 0:
		return fmt.Errorf("invalid MaxContentLength %d: must not be negative", options.MaxContentLength)
	case options.MaxBodyLength < 0:
		return fmt.Errorf("invalid MaxBodyLength %d: must not be negative", options.MaxBodyLength)
	case options.MaxRedirects < 0:
		return fmt.Errorf("invalid MaxRedirects %d: must not be negative", options.MaxRedirects)
	}
	if options.Timeout == 0 {
		options.Timeout = 1000
	}
	if options.MaxContentLength == 0 {
		options.MaxContentLength = 2000
	}
	if options.MaxBodyLength == 0 {
		options.MaxBodyLength = 2000
	}
	if options.ResponseType == "" {
		options.ResponseType = "json"
	}
	if options.ResponseEncoding == "" {
		options.ResponseEncoding = "utf8"
	}
	if options.MaxRedirects == 0 {
		options.MaxRedirects = 21
	}
	if options.Method == "" {
		options.Method = "GET"
	}
	if options.DisableDecompression {
		options.Decompress = false
	} else if !options.Decompress {
		// Decompression historically defaulted to true. The separate disable
		// option makes that default explicit without changing existing callers.
		options.Decompress = true
	}
	upperMethod := strings.ToUpper(options.Method)
	if !validMethods[upperMethod] {
		return fmt.Errorf("invalid HTTP method: %q", options.Method)
	}
	options.Method = upperMethod
	return nil
}

// hasDotDotSegment reports whether a relative URL contains a ".." path
// segment, in raw or percent-encoded form.
func hasDotDotSegment(rel string) bool {
	if i := strings.IndexAny(rel, "?#"); i >= 0 {
		rel = rel[:i]
	}
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".." {
			return true
		}
		if unescaped, err := url.PathUnescape(seg); err == nil && unescaped == ".." {
			return true
		}
	}
	return false
}

// joinBaseURL joins rel onto base and guarantees the result stays under the
// base path, so caller-supplied relative URLs cannot reach sibling endpoints.
func joinBaseURL(base, rel string) (string, error) {
	if hasDotDotSegment(rel) {
		return "", fmt.Errorf("url %q must not contain \"..\" path segments when a base URL is set", rel)
	}
	full, err := url.JoinPath(base, rel)
	if err != nil {
		return "", err
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	fullURL, err := url.Parse(full)
	if err != nil {
		return "", err
	}
	basePath := path.Clean("/" + baseURL.EscapedPath())
	if basePath == "/" {
		return full, nil
	}
	fullPath := fullURL.EscapedPath()
	if fullPath != basePath && !strings.HasPrefix(fullPath, basePath+"/") {
		return "", fmt.Errorf("url %q escapes the base URL path %q", rel, baseURL.Path)
	}
	return full, nil
}

func (c *Client) resolveURL(options *RequestOptions) (string, error) {
	base := c.baseURL()
	if base == "" {
		base = options.BaseURL
	}
	fullURL := options.URL
	if base != "" {
		var err error
		fullURL, err = joinBaseURL(base, options.URL)
		if err != nil {
			return "", err
		}
	}
	if len(options.Params) > 0 {
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			return "", err
		}
		q := parsedURL.Query()
		for k, v := range options.Params {
			q.Add(k, v)
		}
		parsedURL.RawQuery = q.Encode()
		fullURL = parsedURL.String()
	}
	return fullURL, nil
}

// encodeBody serialises the request body and returns the bytes to send.
func encodeBody(body interface{}) ([]byte, error) {
	switch v := body.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return json.Marshal(body)
	}
}

func buildProxyURL(p *Proxy) (*url.URL, error) {
	if !validProxyProtocols[strings.ToLower(p.Protocol)] {
		return nil, fmt.Errorf("invalid proxy protocol %q", p.Protocol)
	}
	if p.Host == "" || strings.ContainsAny(p.Host, "/?#@ \t\r\n\\") {
		return nil, fmt.Errorf("invalid proxy host %q", p.Host)
	}
	if p.Port < 1 || p.Port > 65535 {
		return nil, fmt.Errorf("invalid proxy port %d", p.Port)
	}
	proxyURL := &url.URL{
		Scheme: strings.ToLower(p.Protocol),
		Host:   net.JoinHostPort(p.Host, fmt.Sprint(p.Port)),
	}
	if p.Auth != nil {
		// Credentials in the proxy URL make net/http send Proxy-Authorization
		// for both plain HTTP requests and CONNECT tunnels.
		proxyURL.User = url.UserPassword(p.Auth.Username, p.Auth.Password)
	}
	return proxyURL, nil
}

// proxyTransport returns a cached transport for the given proxy, derived from
// base, so connections are pooled across requests.
func (c *Client) proxyTransport(base http.RoundTripper, p *Proxy) (http.RoundTripper, error) {
	proxyURL, err := buildProxyURL(p)
	if err != nil {
		return nil, err
	}
	baseTransport, ok := base.(*http.Transport)
	if !ok {
		return nil, errors.New("proxy options require an *http.Transport")
	}
	key := fmt.Sprintf("%p|%s", baseTransport, proxyURL.String())
	if cached, ok := c.proxyTransports.Load(key); ok {
		return cached.(*http.Transport), nil
	}
	transport := baseTransport.Clone()
	transport.Proxy = http.ProxyURL(proxyURL)
	actual, loaded := c.proxyTransports.LoadOrStore(key, transport)
	if loaded {
		transport.CloseIdleConnections()
	}
	return actual.(*http.Transport), nil
}

func isDomainOrSubdomain(sub, parent string) bool {
	if sub == parent {
		return true
	}
	return strings.HasSuffix(sub, "."+parent)
}

// stripCredentialsOnRedirect removes every non-allowlisted header when a
// redirect leaves the original host or downgrades from https to http.
func stripCredentialsOnRedirect(req *http.Request, via []*http.Request) {
	if len(via) == 0 {
		return
	}
	initial := via[0].URL
	crossHost := !isDomainOrSubdomain(strings.ToLower(req.URL.Hostname()), strings.ToLower(initial.Hostname()))
	downgrade := strings.EqualFold(initial.Scheme, "https") && !strings.EqualFold(req.URL.Scheme, "https")
	if !crossHost && !downgrade {
		return
	}
	for name := range req.Header {
		if !safeRedirectHeaders[http.CanonicalHeaderKey(name)] {
			req.Header.Del(name)
		}
	}
}

// flattenHeaders converts an http.Header into the map shape used by cache key
// functions.
func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for name, values := range h {
		out[name] = strings.Join(values, ", ")
	}
	return out
}

func responseFromCache(entry *CacheEntry, req *http.Request) *http.Response {
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", entry.StatusCode, http.StatusText(entry.StatusCode)),
		StatusCode:    entry.StatusCode,
		Header:        entry.Headers.Clone(),
		Body:          io.NopCloser(bytes.NewReader(entry.Body)),
		ContentLength: int64(len(entry.Body)),
		Request:       req,
	}
}

func responseForbidsCaching(h http.Header) bool {
	for _, value := range h.Values("Cache-Control") {
		for _, directive := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(directive), "no-store") {
				return true
			}
		}
	}
	return false
}

func (c *Client) Request(options *RequestOptions) (result *Response, err error) {
	if options == nil {
		return nil, errors.New("request options must not be nil")
	}
	// Work on a private copy so the caller's struct and maps are never mutated.
	options = cloneRequestOptions(options)
	if err := applyDefaults(options); err != nil {
		return nil, err
	}

	startTime := time.Now()
	fullURL, err := c.resolveURL(options)
	if err != nil {
		return nil, err
	}

	var bodyBytes []byte
	var bodyReader io.Reader
	if options.Body != nil {
		bodyBytes, err = encodeBody(options.Body)
		if err != nil {
			return nil, err
		}
		if options.MaxBodyLength > 0 && int64(len(bodyBytes)) > options.MaxBodyLength {
			return nil, errors.New("request body length exceeded maxBodyLength")
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(options.Method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}
	if bodyReader != nil && options.OnUploadProgress != nil {
		// Wrap after NewRequest so ContentLength and GetBody stay populated.
		req.Body = io.NopCloser(&ProgressReader{
			reader:     bodyReader,
			total:      int64(len(bodyBytes)),
			onProgress: options.OnUploadProgress,
		})
	}

	for _, interceptor := range options.InterceptorOptions.RequestInterceptors {
		if err := interceptor(req); err != nil {
			return nil, fmt.Errorf("request interceptor failed: %w", err)
		}
	}

	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}
	if options.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if options.Auth != nil {
		auth := options.Auth.Username + ":" + options.Auth.Password
		req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	}
	if !options.Decompress && req.Header.Get("Accept-Encoding") == "" {
		// Setting an explicit encoding prevents net/http from transparently
		// requesting and decompressing gzip responses.
		req.Header.Set("Accept-Encoding", "identity")
	}

	if c.Logger != nil {
		c.Logger.LogRequest(req, options.LogLevel)
	}

	// Cache lookup happens only once the outgoing headers are final, so the key
	// reflects credentials added by interceptors or custom headers.
	var cacheKey string
	shouldCache := shouldCacheRequest(c.CacheConfig, options)
	if shouldCache {
		var bodyHash string
		if len(bodyBytes) > 0 {
			sum := sha256.Sum256(bodyBytes)
			bodyHash = hex.EncodeToString(sum[:])
		}
		cacheKey = generateCacheKey(c.CacheConfig, options, fullURL, flattenHeaders(req.Header), bodyHash)
	}

	if shouldCache && !shouldForceRefresh(options) {
		if entry := c.CacheConfig.Cache.Get(cacheKey); entry != nil {
			cached := responseFromCache(entry, req)
			if c.Logger != nil {
				c.Logger.LogResponse(cached, entry.Body, time.Since(startTime), options.LogLevel)
			}
			if options.ValidateStatus != nil && !options.ValidateStatus(cached.StatusCode) {
				return nil, fmt.Errorf("Request failed with status code: %v", cached.StatusCode)
			}
			for _, interceptor := range options.InterceptorOptions.ResponseInterceptors {
				if err := interceptor(cached); err != nil {
					return nil, fmt.Errorf("response interceptor failed: %w", err)
				}
			}
			return &Response{
				StatusCode: cached.StatusCode,
				Headers:    cached.Header,
				Body:       entry.Body,
			}, nil
		}
	}

	baseHTTPClient := c.HTTPClient
	if baseHTTPClient == nil {
		baseHTTPClient = http.DefaultClient
	}
	httpClient := *baseHTTPClient
	httpClient.Timeout = time.Duration(options.Timeout) * time.Millisecond

	originalCheckRedirect := httpClient.CheckRedirect
	maxRedirects := options.MaxRedirects
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return fmt.Errorf("too many redirects (max: %d)", maxRedirects)
		}
		stripCredentialsOnRedirect(req, via)
		if originalCheckRedirect != nil {
			return originalCheckRedirect(req, via)
		}
		return nil
	}

	if options.Proxy != nil {
		baseTransport := httpClient.Transport
		if baseTransport == nil {
			baseTransport = http.DefaultTransport
		}
		transport, err := c.proxyTransport(baseTransport, options.Proxy)
		if err != nil {
			return nil, err
		}
		httpClient.Transport = transport
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		if c.Logger != nil {
			c.Logger.LogError(err, options.LogLevel)
		}
		return nil, err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close response body: %v", err, cerr)
			} else {
				err = fmt.Errorf("failed to close response body: %w", cerr)
			}
		}
	}()

	var responseBody []byte
	limitedReader := io.LimitReader(resp.Body, options.MaxContentLength+1)

	if options.OnDownloadProgress != nil {
		buf := &bytes.Buffer{}
		progressWriter := &ProgressWriter{
			writer:     buf,
			total:      resp.ContentLength,
			onProgress: options.OnDownloadProgress,
		}
		if _, err := io.Copy(progressWriter, limitedReader); err != nil {
			return nil, err
		}
		responseBody = buf.Bytes()
	} else {
		responseBody, err = io.ReadAll(limitedReader)
		if err != nil {
			return nil, err
		}
	}

	if int64(len(responseBody)) > options.MaxContentLength {
		return nil, errors.New("response content length exceeded maxContentLength")
	}

	if c.Logger != nil {
		c.Logger.LogResponse(resp, responseBody, time.Since(startTime), options.LogLevel)
	}

	if options.ValidateStatus != nil && !(options.ValidateStatus(resp.StatusCode)) {
		return nil, fmt.Errorf("Request failed with status code: %v", resp.StatusCode)
	}

	for _, interceptor := range options.InterceptorOptions.ResponseInterceptors {
		if err := interceptor(resp); err != nil {
			return nil, fmt.Errorf("response interceptor failed: %w", err)
		}
	}

	if shouldCache && resp.StatusCode >= 200 && resp.StatusCode < 300 && !responseForbidsCaching(resp.Header) {
		if ttl := getCacheTTL(c.CacheConfig, options); ttl > 0 {
			headers := resp.Header.Clone()
			// Cookies are per-client state and must never be replayed to
			// another caller from the cache.
			headers.Del("Set-Cookie")
			c.CacheConfig.Cache.Set(cacheKey, &CacheEntry{
				Body:       responseBody,
				StatusCode: resp.StatusCode,
				Headers:    headers,
				CreatedAt:  time.Now(),
			}, ttl)
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}, nil
}

func mergeOptions(dst, src *RequestOptions) {
	if src.Method != "" {
		dst.Method = src.Method
	}
	if src.URL != "" {
		dst.URL = src.URL
	}
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.Params != nil {
		dst.Params = make(map[string]string, len(src.Params))
		for k, v := range src.Params {
			dst.Params[k] = v
		}
	}
	if src.Body != nil {
		dst.Body = src.Body
	}
	if src.Headers != nil {
		dst.Headers = make(map[string]string, len(src.Headers))
		for k, v := range src.Headers {
			dst.Headers[k] = v
		}
	}
	if src.Timeout != 0 {
		dst.Timeout = src.Timeout
	}
	if src.Auth != nil {
		dst.Auth = src.Auth
	}
	if src.ResponseType != "" {
		dst.ResponseType = src.ResponseType
	}
	if src.ResponseEncoding != "" {
		dst.ResponseEncoding = src.ResponseEncoding
	}
	if src.MaxRedirects != 0 {
		dst.MaxRedirects = src.MaxRedirects
	}
	if src.MaxContentLength != 0 {
		dst.MaxContentLength = src.MaxContentLength
	}
	if src.MaxBodyLength != 0 {
		dst.MaxBodyLength = src.MaxBodyLength
	}
	if src.ValidateStatus != nil {
		dst.ValidateStatus = src.ValidateStatus
	}
	if src.InterceptorOptions.RequestInterceptors != nil {
		dst.InterceptorOptions.RequestInterceptors = src.InterceptorOptions.RequestInterceptors
	}
	if src.InterceptorOptions.ResponseInterceptors != nil {
		dst.InterceptorOptions.ResponseInterceptors = src.InterceptorOptions.ResponseInterceptors
	}
	if src.OnUploadProgress != nil {
		dst.OnUploadProgress = src.OnUploadProgress
	}
	if src.OnDownloadProgress != nil {
		dst.OnDownloadProgress = src.OnDownloadProgress
	}
	if src.Proxy != nil {
		dst.Proxy = src.Proxy
	}
	if src.Cache != nil {
		dst.Cache = src.Cache
	}
	if src.Decompress {
		dst.Decompress = true
	}
	dst.DisableDecompression = src.DisableDecompression
}

func SetBaseURL(baseURL string) {
	defaultClient.mu.Lock()
	defer defaultClient.mu.Unlock()
	defaultClient.BaseURL = baseURL
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
		Logger:     NewLogger(LevelNone),
	}
}

// NewClientWithCache creates a client with cache enabled
func NewClientWithCache(baseURL string, cacheConfig *CacheConfig) *Client {
	return &Client{
		BaseURL:     baseURL,
		HTTPClient:  &http.Client{},
		Logger:      NewLogger(LevelNone),
		CacheConfig: cacheConfig,
	}
}

// SetCache sets the cache configuration for the client
func (c *Client) SetCache(config *CacheConfig) {
	c.CacheConfig = config
}

// ClearCache clears all entries in the client's cache
func (c *Client) ClearCache() {
	if c.CacheConfig != nil && c.CacheConfig.Cache != nil {
		c.CacheConfig.Cache.Clear()
	}
}

// CacheStats returns the cache statistics
func (c *Client) CacheStats() *CacheStats {
	if c.CacheConfig != nil && c.CacheConfig.Cache != nil {
		stats := c.CacheConfig.Cache.Stats()
		return &stats
	}
	return nil
}
