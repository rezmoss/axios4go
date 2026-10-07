package axios4go

import (
	"bytes"
	"errors"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Regression tests for the findings of the security audit. Each test documents
// the behaviour the library must guarantee.

func newCacheClient(t *testing.T) *Client {
	t.Helper()
	cache := NewMemoryCache(nil)
	t.Cleanup(cache.Close)
	return NewClientWithCache("", &CacheConfig{Cache: cache, DefaultTTL: time.Minute})
}

func TestSecurity_CacheKeyVariesOnCustomCredentialHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data for " + r.Header.Get("X-Api-Key")))
	}))
	defer server.Close()

	client := newCacheClient(t)
	req := func(key string) *RequestOptions {
		return &RequestOptions{URL: server.URL, Headers: map[string]string{"X-Api-Key": key}, Cache: CacheEnabled(time.Minute)}
	}
	alice, err := client.Request(req("alice"))
	if err != nil {
		t.Fatal(err)
	}
	bob, err := client.Request(req("bob"))
	if err != nil {
		t.Fatal(err)
	}
	if string(alice.Body) != "data for alice" {
		t.Fatalf("unexpected first response %q", alice.Body)
	}
	if string(bob.Body) != "data for bob" {
		t.Fatalf("cache served alice's response to bob: %q", bob.Body)
	}
}

func TestSecurity_CacheKeyIncludesInterceptorHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("data for " + r.Header.Get("Authorization")))
	}))
	defer server.Close()

	client := newCacheClient(t)
	req := func(token string) *RequestOptions {
		return &RequestOptions{
			URL:   server.URL,
			Cache: CacheEnabled(time.Minute),
			InterceptorOptions: InterceptorOptions{RequestInterceptors: RequestInterceptors{
				func(r *http.Request) error { r.Header.Set("Authorization", "Bearer "+token); return nil },
			}},
		}
	}
	if _, err := client.Request(req("alice")); err != nil {
		t.Fatal(err)
	}
	bob, err := client.Request(req("bob"))
	if err != nil {
		t.Fatal(err)
	}
	if string(bob.Body) != "data for Bearer bob" {
		t.Fatalf("cache ignored interceptor-set credentials: %q", bob.Body)
	}
}

func TestSecurity_CacheDoesNotReplaySetCookie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "session=secret")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := newCacheClient(t)
	opts := func() *RequestOptions { return &RequestOptions{URL: server.URL, Cache: CacheEnabled(time.Minute)} }
	if _, err := client.Request(opts()); err != nil {
		t.Fatal(err)
	}
	second, err := client.Request(opts())
	if err != nil {
		t.Fatal(err)
	}
	if second.Headers.Get("Set-Cookie") != "" {
		t.Fatalf("cached response replayed Set-Cookie: %q", second.Headers.Get("Set-Cookie"))
	}
}

func TestSecurity_CacheKeyVariesOnRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_, _ = w.Write([]byte("result for " + string(body)))
	}))
	defer server.Close()

	cache := NewMemoryCache(nil)
	defer cache.Close()
	client := NewClientWithCache("", &CacheConfig{Cache: cache, DefaultTTL: time.Minute, CacheableMethods: []string{"POST"}})
	req := func(q string) *RequestOptions {
		return &RequestOptions{Method: "POST", URL: server.URL, Body: q, Cache: CacheEnabled(time.Minute)}
	}
	if _, err := client.Request(req("query-a")); err != nil {
		t.Fatal(err)
	}
	b, err := client.Request(req("query-b"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b.Body) != "result for query-b" {
		t.Fatalf("cache ignored request body: %q", b.Body)
	}
}

func TestSecurity_CacheHonoursNoStore(t *testing.T) {
	var hits int32
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hits++
		mu.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := newCacheClient(t)
	for i := 0; i < 2; i++ {
		if _, err := client.Request(&RequestOptions{URL: server.URL, Cache: CacheEnabled(time.Minute)}); err != nil {
			t.Fatal(err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if hits != 2 {
		t.Fatalf("response marked no-store was cached: server hits=%d", hits)
	}
}

func TestSecurity_CacheHitRunsValidateStatusAndInterceptors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := newCacheClient(t)
	interceptorCalls := 0
	opts := func(validate func(int) bool) *RequestOptions {
		return &RequestOptions{
			URL:            server.URL,
			Cache:          CacheEnabled(time.Minute),
			ValidateStatus: validate,
			InterceptorOptions: InterceptorOptions{ResponseInterceptors: ResponseInterceptors{
				func(*http.Response) error { interceptorCalls++; return nil },
			}},
		}
	}
	if _, err := client.Request(opts(nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Request(opts(nil)); err != nil {
		t.Fatal(err)
	}
	if interceptorCalls != 2 {
		t.Fatalf("response interceptors skipped on cache hit: calls=%d", interceptorCalls)
	}
	if _, err := client.Request(opts(func(int) bool { return false })); err == nil {
		t.Fatal("ValidateStatus was bypassed on cache hit")
	}
}

func TestSecurity_RequestDoesNotMutateCallerOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))
	defer server.Close()

	headers := map[string]string{"X-A": "1"}
	opts := &RequestOptions{Headers: headers}
	if _, err := Post(server.URL, map[string]string{"k": "v"}, opts); err != nil {
		t.Fatal(err)
	}
	if _, ok := headers["Content-Type"]; ok {
		t.Fatal("library wrote Content-Type into the caller's Headers map")
	}
	if opts.Timeout != 0 || opts.Method != "" || opts.Body != nil {
		t.Fatalf("library mutated the caller's RequestOptions: %+v", opts)
	}

	client := NewClient("")
	direct := &RequestOptions{URL: server.URL, Headers: map[string]string{}}
	if _, err := client.Request(direct); err != nil {
		t.Fatal(err)
	}
	if direct.Timeout != 0 || direct.Method != "" || direct.MaxContentLength != 0 {
		t.Fatalf("Client.Request mutated the caller's RequestOptions: %+v", direct)
	}
}

func TestSecurity_SharedOptionsAreGoroutineSafe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))
	defer server.Close()

	shared := &RequestOptions{Headers: map[string]string{"X-A": "1"}}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Post(server.URL, map[string]string{"k": "v"}, shared); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestSecurity_SetBaseURLIsGoroutineSafe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))
	defer server.Close()
	defer SetBaseURL("")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); SetBaseURL(server.URL) }()
		go func() { defer wg.Done(); _, _ = Get("/x") }()
	}
	wg.Wait()
}

// localhostURL rewrites a 127.0.0.1 httptest URL to use the "localhost" name so
// that net/http treats it as a different host.
func localhostURL(u string) string {
	return strings.Replace(u, "127.0.0.1", "localhost", 1)
}

func TestSecurity_RedirectStripsCredentialsCrossHost(t *testing.T) {
	var got http.Header
	evil := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte("evil"))
	}))
	defer evil.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, localhostURL(evil.URL), http.StatusFound)
	}))
	defer origin.Close()

	_, err := Get(origin.URL, &RequestOptions{Headers: map[string]string{
		"X-Api-Key":     "supersecret",
		"Authorization": "Bearer tok",
		"Accept":        "application/json",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("redirect target never received the request")
	}
	if v := got.Get("X-Api-Key"); v != "" {
		t.Fatalf("custom credential header forwarded to another host: %q", v)
	}
	if v := got.Get("Authorization"); v != "" {
		t.Fatalf("Authorization forwarded to another host: %q", v)
	}
	if got.Get("Accept") != "application/json" {
		t.Fatal("non-sensitive Accept header should still be forwarded")
	}
}

func TestSecurity_RedirectKeepsHeadersSameHost(t *testing.T) {
	var got http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		got = r.Header.Clone()
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	_, err := Get(server.URL+"/start", &RequestOptions{Headers: map[string]string{"X-Api-Key": "k"}, Auth: &Auth{Username: "u", Password: "p"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Get("X-Api-Key") != "k" || got.Get("Authorization") == "" {
		t.Fatalf("same-host redirect lost headers: %v", got)
	}
}

func TestSecurity_RedirectStripsCredentialsOnSchemeDowngrade(t *testing.T) {
	var got http.Header
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte("plain"))
	}))
	defer plain.Close()
	secure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, plain.URL, http.StatusFound)
	}))
	defer secure.Close()

	client := &Client{HTTPClient: secure.Client()}
	_, err := client.Request(&RequestOptions{
		URL:     secure.URL,
		Auth:    &Auth{Username: "u", Password: "p"},
		Headers: map[string]string{"X-Api-Key": "k"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("plain server never received the request")
	}
	if got.Get("Authorization") != "" || got.Get("X-Api-Key") != "" {
		t.Fatalf("credentials sent over plaintext after https->http redirect: %v", got)
	}
}

func TestSecurity_BaseURLRejectsPathTraversal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(r.URL.Path)) }))
	defer server.Close()

	client := NewClient(server.URL + "/v1/tenants/acme/")
	for _, u := range []string{"../../admin/users", "users/../../other", "/..", "a/%2e%2e/%2e%2e/b"} {
		if resp, err := client.Request(&RequestOptions{URL: u}); err == nil {
			t.Errorf("URL %q escaped the base path, server saw %q", u, resp.Body)
		}
	}

	resp, err := client.Request(&RequestOptions{URL: "users/./42"})
	if err != nil {
		t.Fatal(err)
	}
	if string(resp.Body) != "/v1/tenants/acme/users/42" {
		t.Fatalf("legitimate relative URL broken: %q", resp.Body)
	}
}

func startProxy(t *testing.T) (*httptest.Server, func() http.Header) {
	t.Helper()
	var mu sync.Mutex
	var got http.Header
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = r.Header.Clone()
		got.Set("X-Remote-Addr", r.RemoteAddr)
		mu.Unlock()
		_, _ = w.Write([]byte("via proxy"))
	}))
	t.Cleanup(proxy.Close)
	return proxy, func() http.Header { mu.Lock(); defer mu.Unlock(); return got }
}

func proxyFor(t *testing.T, server *httptest.Server, auth *Auth) *Proxy {
	t.Helper()
	host, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	port, _ := strconv.Atoi(portStr)
	return &Proxy{Protocol: "http", Host: host, Port: port, Auth: auth}
}

func TestSecurity_ProxyAuthSentForPlainHTTPTargets(t *testing.T) {
	proxy, got := startProxy(t)
	_, err := Get("http://example.invalid/resource", &RequestOptions{
		Proxy: proxyFor(t, proxy, &Auth{Username: "u", Password: "p"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got() == nil {
		t.Fatal("request never reached proxy")
	}
	if got().Get("Proxy-Authorization") != "Basic dTpw" {
		t.Fatalf("proxy did not receive credentials: %q", got().Get("Proxy-Authorization"))
	}
}

func TestSecurity_ProxyConfigValidation(t *testing.T) {
	bad := []Proxy{
		{Protocol: "http", Host: "user:pw@evil.example", Port: 8080},
		{Protocol: "http", Host: "proxy.example/path", Port: 8080},
		{Protocol: "http", Host: "proxy.example?x=1", Port: 8080},
		{Protocol: "http", Host: "", Port: 8080},
		{Protocol: "http", Host: "proxy.example", Port: 0},
		{Protocol: "http", Host: "proxy.example", Port: 70000},
		{Protocol: "javascript", Host: "proxy.example", Port: 8080},
	}
	for _, p := range bad {
		p := p
		if _, err := Get("http://example.invalid/", &RequestOptions{Proxy: &p}); err == nil {
			t.Errorf("proxy config %+v accepted", p)
		} else if !strings.Contains(err.Error(), "proxy") {
			t.Errorf("proxy config %+v produced non-validation error: %v", p, err)
		}
	}
}

func TestSecurity_ProxyTransportIsReused(t *testing.T) {
	proxy, got := startProxy(t)
	p := proxyFor(t, proxy, &Auth{Username: "u", Password: "p"})
	addrs := map[string]bool{}
	for i := 0; i < 3; i++ {
		if _, err := Get("http://example.invalid/", &RequestOptions{Proxy: p}); err != nil {
			t.Fatal(err)
		}
		addrs[got().Get("X-Remote-Addr")] = true
	}
	if len(addrs) != 1 {
		t.Fatalf("each proxied request opened a new connection (new Transport per request): %v", addrs)
	}
}

func TestSecurity_NegativeLimitsRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))
	defer server.Close()

	cases := map[string]*RequestOptions{
		"Timeout":          {Timeout: -1},
		"MaxBodyLength":    {MaxBodyLength: -1, Body: "x"},
		"MaxContentLength": {MaxContentLength: -1},
		"MaxRedirects":     {MaxRedirects: -1},
	}
	for name, opts := range cases {
		if _, err := Get(server.URL, opts); err == nil {
			t.Errorf("negative %s accepted", name)
		}
	}
}

func TestSecurity_MethodIsNormalised(t *testing.T) {
	var seen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r.Method }))
	defer server.Close()
	if _, err := Request("get", server.URL); err != nil {
		t.Fatal(err)
	}
	if seen != "GET" {
		t.Fatalf("method sent as %q", seen)
	}
}

type closeErrBody struct{ io.Reader }

func (closeErrBody) Close() error { return errors.New("boom") }

type closeErrTransport struct{}

func (closeErrTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: closeErrBody{strings.NewReader("ok")}, Request: r}, nil
}

func TestSecurity_BodyCloseErrorIsReported(t *testing.T) {
	client := &Client{HTTPClient: &http.Client{Transport: closeErrTransport{}}}
	resp, err := client.Request(&RequestOptions{URL: "http://example.invalid/"})
	if err == nil || !strings.Contains(err.Error(), "close") {
		t.Fatalf("body close error dropped: err=%v", err)
	}
	if resp == nil || string(resp.Body) != "ok" {
		t.Fatalf("response should still be returned alongside the close error: %+v", resp)
	}
}

func TestSecurity_PromiseCallbackPanicRoutesToCatchAndUnblocksFinally(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) }))
	defer server.Close()

	var caught error
	done := make(chan struct{})
	go func() {
		defer close(done)
		GetAsync(server.URL).
			Then(func(*Response) { panic("callback exploded") }).
			Catch(func(err error) { caught = err }).
			Finally(func() {})
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Finally never returned after a panicking Then callback")
	}
	if caught == nil || !strings.Contains(caught.Error(), "callback exploded") {
		t.Fatalf("panic was not routed to Catch: %v", caught)
	}
}

func TestSecurity_LoggerRedactsSecrets(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger(LogOptions{Level: LevelDebug, Output: &buf, IncludeHeaders: true,
		MaskHeaders: []string{"Authorization", "Cookie", "Set-Cookie", "Proxy-Authorization", "X-Api-Key"}})

	req, _ := http.NewRequest("GET", "https://user:hunter2@example.com/path", nil)
	req.Header.Set("X-Api-Key", "apikey-secret")
	req.Header.Set("X-Trace", "line1\r\nFAKE-LOG-LINE: injected")
	logger.LogRequest(req, LevelDebug)
	out := buf.String()

	if strings.Contains(out, "hunter2") {
		t.Fatalf("URL password logged: %s", out)
	}
	if strings.Contains(out, "apikey-secret") {
		t.Fatalf("X-Api-Key value logged: %s", out)
	}
	if strings.Contains(out, "\nFAKE-LOG-LINE") {
		t.Fatalf("header value injected a log line: %q", out)
	}

	// The package default logger must mask common credential headers.
	buf.Reset()
	def := NewLogger(LevelDebug).(*DefaultLogger)
	def.options.Output = &buf
	req2, _ := http.NewRequest("GET", "https://example.com/", nil)
	req2.Header.Set("X-Api-Key", "default-secret")
	req2.Header.Set("Proxy-Authorization", "Basic default-proxy-secret")
	def.LogRequest(req2, LevelDebug)
	if strings.Contains(buf.String(), "default-secret") || strings.Contains(buf.String(), "default-proxy-secret") {
		t.Fatalf("default logger leaked credential headers: %s", buf.String())
	}
}

// TestSecurity_AtomicFieldsAligned type-checks the package for 32-bit targets and
// verifies that every int64 used with sync/atomic is 8-byte aligned, since an
// unaligned atomic access panics on 386 and arm.
func TestSecurity_AtomicFieldsAligned(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	var files []*ast.File
	for _, f := range pkgs["axios4go"].Files {
		files = append(files, f)
	}
	for _, arch := range []string{"386", "arm"} {
		sizes := types.SizesFor("gc", arch)
		conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil), Sizes: sizes}
		pkg, err := conf.Check("axios4go", fset, files, nil)
		if err != nil {
			t.Fatalf("%s: %v", arch, err)
		}
		st := pkg.Scope().Lookup("MemoryCache").Type().Underlying().(*types.Struct)
		var fields []*types.Var
		for i := 0; i < st.NumFields(); i++ {
			fields = append(fields, st.Field(i))
		}
		offsets := sizes.Offsetsof(fields)
		for i, f := range fields {
			if f.Name() != "hits" && f.Name() != "misses" {
				continue
			}
			if offsets[i]%8 != 0 {
				t.Errorf("%s: MemoryCache.%s at offset %d is not 8-byte aligned (atomic access would panic)", arch, f.Name(), offsets[i])
			}
		}
	}
}
