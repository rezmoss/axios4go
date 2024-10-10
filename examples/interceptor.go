package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rezmoss/axios4go"
)

func main() {
	// Define request interceptor
	requestInterceptor := func(req *http.Request) error {
		req.Header.Set("User-Agent", "axios4go-example")
		fmt.Println("Request interceptor: Added User-Agent header")
		return nil
	}

	// Define response interceptor
	responseInterceptor := func(resp *http.Response) error {
		fmt.Printf("Response interceptor: Status Code %d\n", resp.StatusCode)
		return nil
	}

	// Create a new client
	client := axios4go.NewClient("https://api.github.com")

	// Create custom transport with interceptors
	transport := &InterceptorTransport{
		Base:                http.DefaultTransport,
		RequestInterceptor:  requestInterceptor,
		ResponseInterceptor: responseInterceptor,
	}

	// Set the custom transport to the client's HTTP client
	client.HTTPClient.Transport = transport

	// Send the request
	response, err := client.Request(&axios4go.RequestOptions{})
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	// Print the response status and body
	fmt.Printf("Response Status: %d\n", response.StatusCode)
	fmt.Printf("Response Body: %s\n", string(response.Body))
}

// InterceptorTransport is a custom http.RoundTripper that applies interceptors
type InterceptorTransport struct {
	Base                http.RoundTripper
	RequestInterceptor  func(*http.Request) error
	ResponseInterceptor func(*http.Response) error
}

func (t *InterceptorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Apply request interceptor
	if t.RequestInterceptor != nil {
		if err := t.RequestInterceptor(req); err != nil {
			return nil, err
		}
	}

	// Perform the actual request
	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Apply response interceptor
	if t.ResponseInterceptor != nil {
		if err := t.ResponseInterceptor(resp); err != nil {
			return nil, err
		}
	}

	return resp, nil
}
