package axios4go

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get":
			json.NewEncoder(w).Encode(map[string]string{"message": "get success"})
		case "/getByProxy":
			json.NewEncoder(w).Encode(map[string]string{"message": "get success by proxy"})
		case "/post":
			json.NewEncoder(w).Encode(map[string]string{"message": "post success"})
		case "/put":
			json.NewEncoder(w).Encode(map[string]string{"message": "put success"})
		case "/delete":
			json.NewEncoder(w).Encode(map[string]string{"message": "delete success"})
		case "/head":
			w.Header().Set("X-Test-Header", "test-value")
		case "/options":
			w.Header().Set("Allow", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
		case "/patch":
			json.NewEncoder(w).Encode(map[string]string{"message": "patch success"})
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestGet(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Simple Style", func(t *testing.T) {
		response, err := Get(server.URL + "/get")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "get success" {
			t.Errorf("Expected message 'get success', got '%s'", result["message"])
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		promise := GetAsync(server.URL + "/get")
		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				// Assertions...
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		response, err := Request("GET", server.URL+"/get")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "get success" {
			t.Errorf("Expected message 'get success', got '%s'", result["message"])
		}
	})
}

func TestPost(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Simple Style", func(t *testing.T) {
		body := map[string]string{"key": "value"}
		response, err := Post(server.URL+"/post", body)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "post success" {
			t.Errorf("Expected message 'post success', got '%s'", result["message"])
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		body := map[string]string{"key": "value"}

		promise := PostAsync(server.URL+"/post", body)

		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				if response.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
				}

				var result map[string]string
				err := json.Unmarshal(response.Body, &result)
				if err != nil {
					t.Errorf("Error unmarshaling response body: %v", err)
				}

				if result["message"] != "post success" {
					t.Errorf("Expected message 'post success', got '%s'", result["message"])
				}
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		body := map[string]string{"key": "value"}
		response, err := Request("POST", server.URL+"/post", &RequestOptions{Body: body})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "post success" {
			t.Errorf("Expected message 'post success', got '%s'", result["message"])
		}
	})
}

func TestPut(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Simple Style", func(t *testing.T) {
		body := map[string]string{"key": "updated_value"}
		response, err := Put(server.URL+"/put", body)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "put success" {
			t.Errorf("Expected message 'put success', got '%s'", result["message"])
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		body := map[string]string{"key": "updated_value"}

		promise := PutAsync(server.URL+"/put", body)

		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				if response.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
				}

				var result map[string]string
				err := json.Unmarshal(response.Body, &result)
				if err != nil {
					t.Errorf("Error unmarshaling response body: %v", err)
				}

				if result["message"] != "put success" {
					t.Errorf("Expected message 'put success', got '%s'", result["message"])
				}
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		body := map[string]string{"key": "updated_value"}
		response, err := Request("PUT", server.URL+"/put", &RequestOptions{Body: body})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "put success" {
			t.Errorf("Expected message 'put success', got '%s'", result["message"])
		}
	})
}

func TestDelete(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Simple Style", func(t *testing.T) {
		response, err := Delete(server.URL + "/delete")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "delete success" {
			t.Errorf("Expected message 'delete success', got '%s'", result["message"])
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		promise := DeleteAsync(server.URL + "/delete")

		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				if response.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
				}

				var result map[string]string
				err := json.Unmarshal(response.Body, &result)
				if err != nil {
					t.Errorf("Error unmarshaling response body: %v", err)
				}

				if result["message"] != "delete success" {
					t.Errorf("Expected message 'delete success', got '%s'", result["message"])
				}
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		response, err := Request("DELETE", server.URL+"/delete")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "delete success" {
			t.Errorf("Expected message 'delete success', got '%s'", result["message"])
		}
	})
}

func TestHead(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Simple Style", func(t *testing.T) {
		response, err := Head(server.URL + "/head")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		if response.Headers.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected X-Test-Header to be 'test-value', got '%s'", response.Headers.Get("X-Test-Header"))
		}

		if len(response.Body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(response.Body))
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		promise := HeadAsync(server.URL + "/head")

		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				if response.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
				}

				if response.Headers.Get("X-Test-Header") != "test-value" {
					t.Errorf("Expected X-Test-Header to be 'test-value', got '%s'", response.Headers.Get("X-Test-Header"))
				}

				if len(response.Body) != 0 {
					t.Errorf("Expected empty body, got %d bytes", len(response.Body))
				}
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		response, err := Request("HEAD", server.URL+"/head")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		if response.Headers.Get("X-Test-Header") != "test-value" {
			t.Errorf("Expected X-Test-Header to be 'test-value', got '%s'", response.Headers.Get("X-Test-Header"))
		}

		if len(response.Body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(response.Body))
		}
	})
}

func TestOptions(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	expectedAllowHeader := "GET, POST, PUT, DELETE, HEAD, OPTIONS"

	t.Run("Simple Style", func(t *testing.T) {
		response, err := Options(server.URL + "/options")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		allowHeader := response.Headers.Get("Allow")
		if allowHeader != expectedAllowHeader {
			t.Errorf("Expected Allow header to be '%s', got '%s'", expectedAllowHeader, allowHeader)
		}

		if len(response.Body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(response.Body))
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		promise := OptionsAsync(server.URL + "/options")

		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				if response.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
				}

				allowHeader := response.Headers.Get("Allow")
				if allowHeader != expectedAllowHeader {
					t.Errorf("Expected Allow header to be '%s', got '%s'", expectedAllowHeader, allowHeader)
				}

				if len(response.Body) != 0 {
					t.Errorf("Expected empty body, got %d bytes", len(response.Body))
				}
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		response, err := Request("OPTIONS", server.URL+"/options")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		allowHeader := response.Headers.Get("Allow")
		if allowHeader != expectedAllowHeader {
			t.Errorf("Expected Allow header to be '%s', got '%s'", expectedAllowHeader, allowHeader)
		}

		if len(response.Body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(response.Body))
		}
	})
}

func TestPatch(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Simple Style", func(t *testing.T) {
		body := map[string]string{"key": "patched_value"}
		response, err := Patch(server.URL+"/patch", body)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "patch success" {
			t.Errorf("Expected message 'patch success', got '%s'", result["message"])
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		body := map[string]string{"key": "patched_value"}

		promise := PatchAsync(server.URL+"/patch", body)

		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				if response.StatusCode != http.StatusOK {
					t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
				}

				var result map[string]string
				err := json.Unmarshal(response.Body, &result)
				if err != nil {
					t.Errorf("Error unmarshaling response body: %v", err)
				}

				if result["message"] != "patch success" {
					t.Errorf("Expected message 'patch success', got '%s'", result["message"])
				}
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		body := map[string]string{"key": "patched_value"}
		response, err := Request("PATCH", server.URL+"/patch", &RequestOptions{Body: body})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response body: %v", err)
		}

		if result["message"] != "patch success" {
			t.Errorf("Expected message 'patch success', got '%s'", result["message"])
		}
	})
}

func TestValidateStatus(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	reqOptions := &RequestOptions{
		ValidateStatus: func(StatusCode int) bool {
			if StatusCode == 200 {
				return false
			}
			return true
		},
	}

	t.Run("Simple Style", func(t *testing.T) {
		response, err := Get(server.URL+"/get", reqOptions)
		if err == nil || response != nil {
			t.Fatalf("Expected error, got %v", err)
		}
		if err.Error() != "Request failed with status code: 200" {
			t.Errorf("Expected error Request failed with status code: 200, got %v", err.Error())
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		promise := GetAsync(server.URL+"/get", reqOptions)

		var catchExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				t.Error("Then should not be executed when validateStatus returns false")
			}).
			Catch(func(err error) {
				if err == nil {
					t.Fatal("Expected an error, got nil")
				}
				expectedError := "Request failed with status code: 200"
				if err.Error() != expectedError {
					t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
				}
				catchExecuted = true
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !catchExecuted {
			t.Error("Catch was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		response, err := Request("GET", server.URL+"/get", reqOptions)
		if err == nil || response != nil {
			t.Fatalf("Expected error, got %v", err)
		}
		if err.Error() != "Request failed with status code: 200" {
			t.Errorf("Expected error Request failed with status code: 200, got %v", err.Error())
		}
	})
}

func TestInterceptors(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	var interceptedRequest *http.Request
	requestInterceptorCalled := false
	requestInterceptor := func(req *http.Request) error {
		req.Header.Set("X-Intercepted", "true")
		interceptedRequest = req
		requestInterceptorCalled = true
		return nil
	}

	responseInterceptor := func(resp *http.Response) error {
		resp.Header.Set("X-Intercepted-Response", "true")
		return nil
	}

	opts := &RequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Params: map[string]string{
			"query": "myQuery",
		},
	}

	opts.InterceptorOptions = InterceptorOptions{
		RequestInterceptors:  []func(*http.Request) error{requestInterceptor},
		ResponseInterceptors: []func(*http.Response) error{responseInterceptor},
	}

	t.Run("Interceptors Test", func(t *testing.T) {
		response, err := Get(server.URL+"/get", opts)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !requestInterceptorCalled {
			t.Error("Request interceptor was not called")
		}

		if interceptedRequest != nil {
			if interceptedRequest.Header.Get("X-Intercepted") != "true" {
				t.Errorf("Expected request header 'X-Intercepted' to be 'true', got '%s'", interceptedRequest.Header.Get("X-Intercepted"))
			}
		} else {
			t.Error("Intercepted request is nil")
		}

		if response.Headers.Get("X-Intercepted-Response") != "true" {
			t.Errorf("Expected response header 'X-Intercepted-Response' to be 'true', got '%s'", response.Headers.Get("X-Intercepted-Response"))
		}
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	request, err := http.NewRequest(r.Method, r.RequestURI, nil)
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return
	}
	bytes := make([]byte, response.ContentLength)
	response.Body.Read(bytes)
	w.Write(bytes)
}

func TestGetByProxy(t *testing.T) {
	server := setupTestServer()
	defer server.Close()
	path := "/getByProxy"
	http.HandleFunc("/", handler)
	go func() {
		// start to mock a proxy server
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			return
		}
	}()
	t.Run("Simple Style", func(t *testing.T) {
		response, err := Get(server.URL+path,
			&RequestOptions{
				Proxy: &Proxy{
					Protocol: "http",
					Host:     "localhost",
					Port:     8080,
					Auth: &Auth{
						Username: "username",
						Password: "password",
					},
				},
			},
		)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response Body: %v", err)
		}

		if result["message"] != "get success by proxy" {
			t.Errorf("Expected message 'get success by proxy', got '%s'", result["message"])
		}
	})

	t.Run("Promise Style", func(t *testing.T) {
		promise := GetAsync(server.URL + path)
		var thenExecuted, finallyExecuted bool

		promise.
			Then(func(response *Response) {
				// Assertions...
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

		// Wait for the promise to complete
		<-promise.done

		if !thenExecuted {
			t.Error("Then was not executed")
		}
		if !finallyExecuted {
			t.Error("Finally was not executed")
		}
	})

	t.Run("Request Style", func(t *testing.T) {
		response, err := Request("GET", server.URL+path)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

		var result map[string]string
		err = json.Unmarshal(response.Body, &result)
		if err != nil {
			t.Fatalf("Error unmarshaling response Body: %v", err)
		}

		if result["message"] != "get success by proxy" {
			t.Errorf("Expected message 'get success by proxy', got '%s'", result["message"])
		}
	})
}

func TestProgressCallbacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the request body to trigger upload progress
		_, err := io.Copy(io.Discard, r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		// Simulate a large file for download
		w.Header().Set("Content-Length", "1000000")
		for i := 0; i < 1000000; i++ {
			_, err := w.Write([]byte("a"))
			if err != nil {
				t.Fatalf("Failed to write response: %v", err)
			}
		}
	}))
	defer server.Close()

	uploadCalled := false
	downloadCalled := false

	body := bytes.NewReader([]byte(strings.Repeat("b", 500000))) // 500KB upload

	_, err := Post(server.URL, body, &RequestOptions{
		OnUploadProgress: func(bytesRead, totalBytes int64) {
			uploadCalled = true
			if bytesRead > totalBytes {
				t.Errorf("Upload progress: bytesRead (%d) > totalBytes (%d)", bytesRead, totalBytes)
			}
		},
		OnDownloadProgress: func(bytesRead, totalBytes int64) {
			downloadCalled = true
			if bytesRead > totalBytes {
				t.Errorf("Download progress: bytesRead (%d) > totalBytes (%d)", bytesRead, totalBytes)
			}
		},
		MaxContentLength: 2000000, // Set this to allow our 1MB response
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !uploadCalled {
		t.Error("Upload progress callback was not called")
	}
	if !downloadCalled {
		t.Error("Download progress callback was not called")
	}
}

func TestLogging(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	t.Run("Test Logger Integration", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewDefaultLogger(LogOptions{
			Level:          LevelDebug,
			Output:         &buf,
			IncludeBody:    true,
			IncludeHeaders: true,
			MaskHeaders:    []string{"Authorization"},
		})

		client := &Client{
			HTTPClient: &http.Client{},
			Logger:     logger,
		}

		// Test with sensitive headers
		reqOptions := &RequestOptions{
			Method:   "POST",
			URL:      server.URL + "/post",
			LogLevel: LevelDebug,
			Headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"X-Test":        "test-value",
			},
			Body:             map[string]string{"test": "data"},
			MaxContentLength: 2000, // Add this line
		}

		_, err := client.Request(reqOptions)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		logOutput := buf.String()

		// Verify request logging
		if !strings.Contains(logOutput, "REQUEST: POST") {
			t.Error("Log should contain request method")
		}
		if !strings.Contains(logOutput, "[MASKED]") {
			t.Error("Authorization header should be masked")
		}
		if !strings.Contains(logOutput, "test-value") {
			t.Error("Non-sensitive header should be visible")
		}
		if !strings.Contains(logOutput, "test") {
			t.Error("Request body should be logged")
		}

		// Verify response logging
		if !strings.Contains(logOutput, "RESPONSE: 200") {
			t.Error("Log should contain response status")
		}
		if !strings.Contains(logOutput, "post success") {
			t.Error("Response body should be logged")
		}
	})

	t.Run("Test Log Levels", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewDefaultLogger(LogOptions{
			Level:  LevelError,
			Output: &buf,
		})

		client := &Client{
			HTTPClient: &http.Client{},
			Logger:     logger,
		}

		// Debug level request should not be logged when logger is at Error level
		_, err := client.Request(&RequestOptions{
			Method:           "GET",
			URL:              server.URL + "/get",
			LogLevel:         LevelDebug,
			MaxContentLength: 2000, // Add this line
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if buf.Len() > 0 {
			t.Error("Debug level request should not be logged when logger is at Error level")
		}
	})
}
