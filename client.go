package axios4go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
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
	mu       sync.Mutex
}

type RequestInterceptors []func(*http.Request) error
type ResponseInterceptors []func(*http.Response) error
type InterceptorOptions struct {
	RequestInterceptors  RequestInterceptors
	ResponseInterceptors ResponseInterceptors
}

type RequestOptions struct {
	Method             string
	Url                string
	BaseURL            string
	Params             map[string]string
	Body               interface{}
	Headers            map[string]string
	Timeout            int
	Auth               *Auth
	ResponseType       string
	ResponseEncoding   string
	MaxRedirects       int
	MaxContentLength   int
	MaxBodyLength      int
	Decompress         bool
	ValidateStatus     func(int) bool
	InterceptorOptions InterceptorOptions
	Proxy              *Proxy
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

var defaultClient = &Client{HTTPClient: &http.Client{}}

func (r *Response) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

func (p *Promise) Then(fn func(*Response)) *Promise {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.response != nil && p.err == nil {
		fn(p.response)
	} else {
		p.then = fn
	}
	return p
}

func (p *Promise) Catch(fn func(error)) *Promise {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		fn(p.err)
	} else {
		p.catch = fn
	}
	return p
}

func (p *Promise) Finally(fn func()) {
	p.mu.Lock()

	if p.response != nil || p.err != nil {
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
	defer p.mu.Unlock()

	p.response = resp
	p.err = err

	if p.then != nil && err == nil {
		p.then(resp)
	}
	if p.catch != nil && err != nil {
		p.catch(err)
	}
	if p.finally != nil {
		p.finally()
	}

	close(p.done)
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
		Url:              urlStr,
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
	resp, err := Request(method, urlStr, options...)
	return &Promise{response: resp, err: err}
}

func (c *Client) Request(options *RequestOptions) (*Response, error) {
	var fullURL string
	if c.BaseURL != "" {
		var err error
		fullURL, err = url.JoinPath(c.BaseURL, options.Url)
		if err != nil {
			return nil, err
		}
	} else if options.BaseURL != "" {
		var err error
		fullURL, err = url.JoinPath(options.BaseURL, options.Url)
		if err != nil {
			return nil, err
		}
	} else {
		fullURL = options.Url
	}

	if len(options.Params) > 0 {
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		q := parsedURL.Query()
		for k, v := range options.Params {
			q.Add(k, v)
		}
		parsedURL.RawQuery = q.Encode()
		fullURL = parsedURL.String()
	}

	var bodyReader io.Reader
	var bodyLength int64

	if options.Body != nil {
		switch v := options.Body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
			bodyLength = int64(len(v))
		case []byte:
			bodyReader = bytes.NewReader(v)
			bodyLength = int64(len(v))
		default:
			jsonBody, err := json.Marshal(options.Body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewBuffer(jsonBody)
			bodyLength = int64(len(jsonBody))
		}
		if options.MaxBodyLength > 0 && bodyLength > int64(options.MaxBodyLength) {
			return nil, errors.New("request body length exceeded maxBodyLength")
		}
	}

	req, err := http.NewRequest(options.Method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	for _, interceptor := range options.InterceptorOptions.RequestInterceptors {
		err = interceptor(req)
		if err != nil {
			return nil, fmt.Errorf("request interceptor failed: %w", err)
		}
	}

	if options.Headers == nil {
		options.Headers = make(map[string]string)
	}

	if options.Body != nil {
		if _, exists := options.Headers["Content-Type"]; !exists {
			options.Headers["Content-Type"] = "application/json"
		}
	}

	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	if options.Auth != nil {
		auth := options.Auth.Username + ":" + options.Auth.Password
		basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", "Basic "+basicAuth)
	}

	c.HTTPClient.Timeout = time.Duration(options.Timeout) * time.Millisecond

	if options.MaxRedirects > 0 {
		c.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= options.MaxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		}
	}

	if options.Proxy != nil {
		// support http and https
		proxyStr := fmt.Sprintf("%s://%s:%d", options.Proxy.Protocol, options.Proxy.Host, options.Proxy.Port)
		proxyURL, err := url.Parse(proxyStr)
		if err != nil {
			return nil, err
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		if options.Proxy.Auth != nil {
			auth := options.Proxy.Auth.Username + ":" + options.Proxy.Auth.Password
			basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
			transport.ProxyConnectHeader = http.Header{
				"Proxy-Authorization": {"Basic " + basicAuth},
			}
		}
		c.HTTPClient.Transport = transport
		// cancel proxy after request
		defer func() {
			c.HTTPClient.Transport = nil
		}()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close response body: %v", err, cerr)
			} else {
				err = fmt.Errorf("failed to close response body: %v", cerr)
			}
		}
	}()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if int64(len(responseBody)) > int64(options.MaxContentLength) {
		return nil, errors.New("response content length exceeded maxContentLength")
	}

	if options.ValidateStatus != nil && !(options.ValidateStatus(resp.StatusCode)) {
		return nil, fmt.Errorf("Request failed with status code: %v", resp.StatusCode)
	}

	for _, interceptor := range options.InterceptorOptions.ResponseInterceptors {
		err = interceptor(resp)
		if err != nil {
			return nil, fmt.Errorf("response interceptor failed: %w", err)
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}, err
}

func mergeOptions(dst, src *RequestOptions) {
	if src.Method != "" {
		dst.Method = src.Method
	}
	if src.Url != "" {
		dst.Url = src.Url
	}
	if src.BaseURL != "" {
		dst.BaseURL = src.BaseURL
	}
	if src.Params != nil {
		dst.Params = src.Params
	}
	if src.Body != nil {
		dst.Body = src.Body
	}
	if src.Headers != nil {
		dst.Headers = src.Headers
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
	if src.Proxy != nil {
		dst.Proxy = src.Proxy
	}
	dst.Decompress = src.Decompress
}

func SetBaseURL(baseURL string) {
	defaultClient.BaseURL = baseURL
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}
