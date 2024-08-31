# axios4go

[![Go Reference](https://pkg.go.dev/badge/github.com/rezmoss/axios4go.svg)](https://pkg.go.dev/github.com/rezmoss/axios4go)
[![Go Report Card](https://goreportcard.com/badge/github.com/rezmoss/axios4go)](https://goreportcard.com/report/github.com/rezmoss/axios4go)
[![Release](https://img.shields.io/github/v/release/rezmoss/axios4go.svg?style=flat-square)](https://github.com/rezmoss/axios4go/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

axios4go is a Go HTTP client library inspired by Axios, providing a simple and intuitive API for making HTTP requests. It offers features like JSON handling, configurable instances, and support for various HTTP methods.

## Features

- Simple and intuitive API
- Support for GET, POST, PUT, DELETE, HEAD, OPTIONS, and PATCH methods
- JSON request and response handling
- Configurable client instances
- Timeout and redirect management
- Basic authentication support
- Customizable request options
- Promise-like asynchronous requests

## Installation

To install axios4go, use `go get`:

```bash
go get github.com/rezmoss/axios4go
```

## Usage

### Making a Simple Request

```go
package main

import (
    "fmt"
    "github.com/rezmoss/axios4go"
)

func main() {
    resp, err := axios4go.Get("https://api.example.com/data")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Status Code: %d\n", resp.StatusCode)
    fmt.Printf("Body: %s\n", string(resp.Body))
}
```

### Using Request Options

```go
resp, err := axios4go.Get("https://api.example.com/data", &axios4go.requestOptions{
    timeout: 5000,  // 5 seconds
    headers: map[string]string{
        "Authorization": "Bearer token",
    },
})
```

### Making POST Requests

```go
body := map[string]interface{}{
    "name": "John Doe",
    "age":  30,
}
resp, err := axios4go.Post("https://api.example.com/users", body)
```

### Using Async Requests

```go
axios4go.GetAsync("https://api.example.com/data").
    Then(func(response *axios4go.Response) {
        fmt.Printf("Status Code: %d\n", response.StatusCode)
        fmt.Printf("Body: %s\n", string(response.Body))
    }).
    Catch(func(err error) {
        fmt.Printf("Error: %v\n", err)
    }).
    Finally(func() {
        fmt.Println("Request completed")
    })
```

### Creating a Custom Client

```go
client := axios4go.NewClient("https://api.example.com")
resp, err := client.Request(&axios4go.requestOptions{
    method: "GET",
    url:    "/users",
})
```

## Configuration Options

axios4go supports various configuration options through the `requestOptions` struct:

- `method`: HTTP method (GET, POST, etc.)
- `url`: Request URL
- `baseURL`: Base URL for the request
- `params`: URL parameters
- `body`: Request body
- `headers`: Custom headers
- `timeout`: Request timeout in milliseconds
- `auth`: Basic authentication credentials
- `maxRedirects`: Maximum number of redirects to follow
- `maxContentLength`: Maximum allowed response content length
- `maxBodyLength`: Maximum allowed request body length

## Contributing

Contributions to axios4go are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.