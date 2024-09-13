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

type requestOptions struct {
	method           string
	url              string
	baseURL          string
	params           map[string]string
	body             interface{}
	headers          map[string]string
	timeout          int
	auth             *auth
	responseType     string
	responseEncoding string
	maxRedirects     int
	maxContentLength int
	maxBodyLength    int
	decompress       bool
	validateStatus   func(int) bool
}

type auth struct {
	username string
	password string
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

func Get(urlStr string, options ...*requestOptions) (*Response, error) {
	return Request("GET", urlStr, options...)
}

func GetAsync(urlStr string, options ...*requestOptions) *Promise {
	promise := NewPromise()

	go func() {
		resp, err := Request("GET", urlStr, options...)
		promise.resolve(resp, err)
	}()

	return promise
}

func Post(urlStr string, body interface{}, options ...*requestOptions) (*Response, error) {
	mergedOptions := mergeBodyIntoOptions(body, options)
	return Request("POST", urlStr, mergedOptions)
}

func PostAsync(urlStr string, body interface{}, options ...*requestOptions) *Promise {
	mergedOptions := mergeBodyIntoOptions(body, options)
	promise := NewPromise()

	go func() {
		resp, err := Request("POST", urlStr, mergedOptions)
		promise.resolve(resp, err)
	}()

	return promise
}

func mergeBodyIntoOptions(body interface{}, options []*requestOptions) *requestOptions {
	mergedOption := &requestOptions{
		body: body,
	}

	if len(options) > 0 {
		*mergedOption = *options[0]
		mergedOption.body = body
	}

	return mergedOption
}

func Put(urlStr string, body interface{}, options ...*requestOptions) (*Response, error) {
	mergedOptions := mergeBodyIntoOptions(body, options)
	return Request("PUT", urlStr, mergedOptions)
}

func PutAsync(urlStr string, body interface{}, options ...*requestOptions) *Promise {
	mergedOptions := mergeBodyIntoOptions(body, options)
	promise := NewPromise()

	go func() {
		resp, err := Request("PUT", urlStr, mergedOptions)
		promise.resolve(resp, err)
	}()

	return promise
}

func Delete(urlStr string, options ...*requestOptions) (*Response, error) {
	return Request("DELETE", urlStr, options...)
}

func DeleteAsync(urlStr string, options ...*requestOptions) *Promise {
	promise := NewPromise()
	go func() {
		resp, err := Request("DELETE", urlStr, options...)
		promise.resolve(resp, err)
	}()
	return promise
}

func Head(urlStr string, options ...*requestOptions) (*Response, error) {
	return Request("HEAD", urlStr, options...)
}

func HeadAsync(urlStr string, options ...*requestOptions) *Promise {
	promise := NewPromise()
	go func() {
		resp, err := Request("HEAD", urlStr, options...)
		promise.resolve(resp, err)
	}()
	return promise
}

func Options(urlStr string, options ...*requestOptions) (*Response, error) {
	return Request("OPTIONS", urlStr, options...)
}

func OptionsAsync(urlStr string, options ...*requestOptions) *Promise {
	promise := NewPromise()
	go func() {
		resp, err := Request("OPTIONS", urlStr, options...)
		promise.resolve(resp, err)
	}()
	return promise
}

func Patch(urlStr string, body interface{}, options ...*requestOptions) (*Response, error) {
	mergedOptions := mergeBodyIntoOptions(body, options)
	return Request("PATCH", urlStr, mergedOptions)
}

func PatchAsync(urlStr string, body interface{}, options ...*requestOptions) *Promise {
	mergedOptions := mergeBodyIntoOptions(body, options)
	promise := NewPromise()

	go func() {
		resp, err := Request("PATCH", urlStr, mergedOptions)
		promise.resolve(resp, err)
	}()

	return promise
}

func Request(method, urlStr string, options ...*requestOptions) (*Response, error) {
	reqOptions := &requestOptions{
		method:           "GET",
		url:              urlStr,
		timeout:          1000,
		responseType:     "json",
		responseEncoding: "utf8",
		maxContentLength: 2000,
		maxBodyLength:    2000,
		maxRedirects:     21,
		decompress:       true,
		validateStatus:   nil,
	}

	if len(options) > 0 && options[0] != nil {
		mergeOptions(reqOptions, options[0])
	}

	if method != "" {
		reqOptions.method = method
	}

	return defaultClient.Request(reqOptions)
}

func RequestAsync(method, urlStr string, options ...*requestOptions) *Promise {
	resp, err := Request(method, urlStr, options...)
	return &Promise{response: resp, err: err}
}

func (c *Client) Request(options *requestOptions) (*Response, error) {
	var fullURL string
	if c.BaseURL != "" {
		var err error
		fullURL, err = url.JoinPath(c.BaseURL, options.url)
		if err != nil {
			return nil, err
		}
	} else if options.baseURL != "" {
		var err error
		fullURL, err = url.JoinPath(options.baseURL, options.url)
		if err != nil {
			return nil, err
		}
	} else {
		fullURL = options.url
	}

	if len(options.params) > 0 {
		parsedURL, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		q := parsedURL.Query()
		for k, v := range options.params {
			q.Add(k, v)
		}
		parsedURL.RawQuery = q.Encode()
		fullURL = parsedURL.String()
	}

	var bodyReader io.Reader
	var bodyLength int64

	if options.body != nil {
		switch v := options.body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
			bodyLength = int64(len(v))
		case []byte:
			bodyReader = bytes.NewReader(v)
			bodyLength = int64(len(v))
		default:
			jsonBody, err := json.Marshal(options.body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewBuffer(jsonBody)
			bodyLength = int64(len(jsonBody))
		}
		if options.maxBodyLength > 0 && bodyLength > int64(options.maxBodyLength) {
			return nil, errors.New("request body length exceeded maxBodyLength")
		}
	}

	req, err := http.NewRequest(options.method, fullURL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range options.headers {
		req.Header.Set(key, value)
	}

	if options.auth != nil {
		auth := options.auth.username + ":" + options.auth.password
		basicAuth := base64.StdEncoding.EncodeToString([]byte(auth))
		req.Header.Set("Authorization", "Basic "+basicAuth)
	}

	c.HTTPClient.Timeout = time.Duration(options.timeout) * time.Millisecond

	if options.maxRedirects > 0 {
		c.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= options.maxRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		}
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

	if int64(len(responseBody)) > int64(options.maxContentLength) {
		return nil, errors.New("response content length exceeded maxContentLength")
	}

	if options.validateStatus != nil && !(options.validateStatus(resp.StatusCode)) {
		return nil, fmt.Errorf("Request failed with status code: %v", resp.StatusCode)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       responseBody,
	}, err
}

func mergeOptions(dst, src *requestOptions) {
	if src.method != "" {
		dst.method = src.method
	}
	if src.url != "" {
		dst.url = src.url
	}
	if src.baseURL != "" {
		dst.baseURL = src.baseURL
	}
	if src.params != nil {
		dst.params = src.params
	}
	if src.body != nil {
		dst.body = src.body
	}
	if src.headers != nil {
		dst.headers = src.headers
	}
	if src.timeout != 0 {
		dst.timeout = src.timeout
	}
	if src.auth != nil {
		dst.auth = src.auth
	}
	if src.responseType != "" {
		dst.responseType = src.responseType
	}
	if src.responseEncoding != "" {
		dst.responseEncoding = src.responseEncoding
	}
	if src.maxRedirects != 0 {
		dst.maxRedirects = src.maxRedirects
	}
	if src.maxContentLength != 0 {
		dst.maxContentLength = src.maxContentLength
	}
	if src.maxBodyLength != 0 {
		dst.maxBodyLength = src.maxBodyLength
	}
	if src.validateStatus != nil {
		dst.validateStatus = src.validateStatus
	}
	dst.decompress = src.decompress
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
