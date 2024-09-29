# axios4go

[![Go Reference](https://pkg.go.dev/badge/github.com/rezmoss/axios4go.svg)](https://pkg.go.dev/github.com/rezmoss/axios4go)
[![Go Report Card](https://goreportcard.com/badge/github.com/rezmoss/axios4go)](https://goreportcard.com/report/github.com/rezmoss/axios4go)
[![Release](https://img.shields.io/github/v/release/rezmoss/axios4go.svg?style=flat-square)](https://github.com/rezmoss/axios4go/releases)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/9401/badge)](https://www.bestpractices.dev/projects/9401)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

axios4go is a Go HTTP client library inspired by Axios, providing a simple and intuitive API for making HTTP requests. It offers features like JSON handling, configurable instances, and support for various HTTP methods.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Making a Simple Request](#making-a-simple-request)
  - [Using Request Options](#using-request-options)
  - [Making POST Requests](#making-post-requests)
  - [Using Async Requests](#using-async-requests)
  - [Creating a Custom Client](#creating-a-custom-client)
- [Configuration Options](#configuration-options)
- [Contributing](#contributing)
- [License](#license)

## Features

- Simple and intuitive API
- Support for `GET`, `POST`, `PUT`, `DELETE`, `HEAD`, `OPTIONS`, and `PATCH` methods
- JSON request and response handling
- Configurable client instances
- Global timeout and redirect management
- Basic authentication support
- Customizable request options
- Promise-like asynchronous requests

## Installation

To install `axios4go`, use `go get`:

```bash
go get -u github.com/rezmoss/axios4go
```

**Note**: Requires Go 1.13 or later.

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
resp, err := axios4go.Get("https://api.example.com/data", &axios4go.RequestOptions{
    Headers: map[string]string{
        "Authorization": "Bearer token",
    },
    Params: map[string]string{
        "query": "golang",
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
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}
fmt.Printf("Status Code: %d\n", resp.StatusCode)
fmt.Printf("Body: %s\n", string(resp.Body))
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
client := axios4go.NewClient("https://api.example.com", 5*time.Second, 5) // Base URL, Timeout, Max Redirects

resp, err := client.Request(&axios4go.RequestOptions{
    Method: "GET",
    URL:    "/users",
    Headers: map[string]string{
        "Authorization": "Bearer token",
    },
})
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}
fmt.Printf("Status Code: %d\n", resp.StatusCode)
fmt.Printf("Body: %s\n", string(resp.Body))
```

**Note**: When creating a custom client, timeouts and max redirects are set at the client level.

## Configuration Options

`axios4go` supports various configuration options through the `RequestOptions` struct:

- **Method**: HTTP method (`GET`, `POST`, etc.)
- **URL**: Request URL (relative to `BaseURL` if provided)
- **BaseURL**: Base URL for the request (overrides client's `BaseURL` if set)
- **Params**: URL query parameters (`map[string]string`)
- **Body**: Request body (can be `string`, `[]byte`, or any JSON serializable object)
- **Headers**: Custom headers (`map[string]string`)
- **Auth**: Basic authentication credentials (`&Auth{Username: "user", Password: "pass"}`)
- **ValidateStatus**: Function to validate HTTP response status codes

**Example**:

```go
options := &axios4go.RequestOptions{
    Method: "POST",
    URL:    "/submit",
    Headers: map[string]string{
        "Content-Type": "application/json",
    },
    Body: map[string]interface{}{
        "title":   "Sample",
        "content": "This is a sample post.",
    },
    Auth: &axios4go.Auth{
        Username: "user",
        Password: "pass",
    },
    Params: map[string]string{
        "verbose": "true",
    },
    ValidateStatus: func(statusCode int) bool {
        return statusCode >= 200 && statusCode < 300
    },
}

resp, err := client.Request(options)
```

## Contributing

Contributions to `axios4go` are welcome! Please follow these guidelines:

- **Fork the repository** and create a new branch for your feature or bug fix.
- **Ensure your code follows Go conventions** and passes all tests.
- **Write tests** for new features or bug fixes.
- **Submit a Pull Request** with a clear description of your changes.

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.