package axios4go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
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
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

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
				thenExecuted = true
			}).
			Catch(func(err error) {
				t.Errorf("Expected no error, got %v", err)
			}).
			Finally(func() {
				finallyExecuted = true
			})

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
		_, err := io.Copy(io.Discard, r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

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
			MaxContentLength: 2000,
		}

		_, err := client.Request(reqOptions)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		logOutput := buf.String()

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

		_, err := client.Request(&RequestOptions{
			Method:           "GET",
			URL:              server.URL + "/get",
			LogLevel:         LevelDebug,
			MaxContentLength: 2000,
		})
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if buf.Len() > 0 {
			t.Error("Debug level request should not be logged when logger is at Error level")
		}
	})
}

func TestTimeoutHandling(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "slow response"}`))
	}))
	defer slowServer.Close()

	start := time.Now()
	response, err := Get(slowServer.URL, &RequestOptions{
		Timeout: 1000,
	})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("Expected a timeout error, but got no error and response: %v", response)
	}

	if elapsed >= 2*time.Second {
		t.Errorf("Expected to fail before 2 seconds, but took %v", elapsed)
	}

	t.Logf("Timeout test passed with error: %v", err)
}

func TestMaxRedirects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect1":
			// Return a 301 or 302 redirect to /redirect2
			http.Redirect(w, r, "/redirect2", http.StatusFound)
		case "/redirect2":
			// Return a 301 or 302 redirect to /final
			http.Redirect(w, r, "/final", http.StatusFound)
		case "/final":
			// Final destination
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"final destination"}`))
		default:
			// Return 404 if the path is not one of the above
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Run("FollowRedirects", func(t *testing.T) {
		resp, err := Get(server.URL+"/redirect1", &RequestOptions{
			MaxRedirects: 5,
		})
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}
		var result map[string]string
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			t.Fatalf("Error parsing final JSON: %v", err)
		}
		if result["message"] != "final destination" {
			t.Errorf(`Expected "final destination", got %q`, result["message"])
		}
	})

	t.Run("TooManyRedirects", func(t *testing.T) {
		resp, err := Get(server.URL+"/redirect1", &RequestOptions{
			MaxRedirects: 1,
		})
		if err == nil {
			t.Fatalf("Expected error due to too many redirects, but got response: %v", resp)
		}
		t.Logf("Redirect test got error as expected: %v", err)
	})
}

func TestBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.Path))
	}))
	defer server.Close()

	t.Run("Using DefaultClient with SetBaseURL", func(t *testing.T) {
		SetBaseURL(server.URL + "/api")
		defer SetBaseURL("")

		resp, err := Get("/testBaseUrl")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		if string(resp.Body) != "/api/testBaseUrl" {
			t.Errorf("Expected path '/api/testBaseUrl', got %q", string(resp.Body))
		}
	})

	t.Run("Using NewClient with BaseURL", func(t *testing.T) {
		client := NewClient(server.URL + "/prefix")

		resp, err := client.Request(&RequestOptions{
			Method: "GET",
			URL:    "hello",
		})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		if string(resp.Body) != "/prefix/hello" {
			t.Errorf("Expected path '/prefix/hello', got %q", string(resp.Body))
		}
	})

	t.Run("TrailingSlashInBaseURL", func(t *testing.T) {
		client := NewClient(server.URL + "/api/")

		resp, err := client.Request(&RequestOptions{
			Method: "GET",
			URL:    "/user/123",
		})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		if string(resp.Body) != "/api/user/123" {
			t.Errorf("Expected path '/api/user/123', got %q", string(resp.Body))
		}
	})
}

func TestBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		//  "user:pass" => "Basic dXNlcjpwYXNz"
		expected := "Basic dXNlcjpwYXNz"
		if auth == expected {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"authorized"}`))
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	t.Run("Correct Credentials", func(t *testing.T) {
		opts := &RequestOptions{
			Auth: &Auth{
				Username: "user",
				Password: "pass",
			},
		}
		resp, err := Get(server.URL, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		if string(resp.Body) != `{"status":"authorized"}` {
			t.Errorf("Expected body to be {\"status\":\"authorized\"}, got %s", resp.Body)
		}
	})

	t.Run("Incorrect Credentials", func(t *testing.T) {
		opts := &RequestOptions{
			Auth: &Auth{
				Username: "baduser",
				Password: "wrongpass",
			},
		}
		resp, err := Get(server.URL, opts)
		if err != nil {
			t.Fatalf("Did not expect a transport error, got: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected 401, got %d", resp.StatusCode)
		}
		if !bytes.Contains(resp.Body, []byte("Invalid credentials")) {
			t.Errorf("Expected response body to contain 'Invalid credentials', got %s", resp.Body)
		}
	})

	t.Run("Missing Credentials", func(t *testing.T) {
		resp, err := Get(server.URL)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Expected 401, got %d", resp.StatusCode)
		}
		if !bytes.Contains(resp.Body, []byte("Missing Authorization header")) {
			t.Errorf("Expected response body to contain 'Missing Authorization header', got %s", resp.Body)
		}
	})
}

func TestParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.URL.RawQuery))
	}))
	defer server.Close()

	t.Run("Single Param", func(t *testing.T) {
		opts := &RequestOptions{
			Params: map[string]string{
				"foo": "bar",
			},
		}
		resp, err := Get(server.URL, opts)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		if string(resp.Body) != "foo=bar" {
			t.Errorf("Expected 'foo=bar', got %q", resp.Body)
		}
	})

	t.Run("Multiple Params", func(t *testing.T) {
		opts := &RequestOptions{
			Params: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
		}
		resp, err := Get(server.URL, opts)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}

		rawQuery := string(resp.Body)
		if !strings.Contains(rawQuery, "param1=value1") ||
			!strings.Contains(rawQuery, "param2=value2") {
			t.Errorf("Expected query to contain param1=value1 and param2=value2, got %q", rawQuery)
		}
	})

	t.Run("Special Characters", func(t *testing.T) {
		opts := &RequestOptions{
			Params: map[string]string{
				"q": "hello world",
			},
		}
		resp, err := Get(server.URL, opts)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		rawQuery := string(resp.Body)
		if !strings.Contains(rawQuery, "q=hello") {
			t.Errorf("Expected query to contain 'q=hello', got %q", rawQuery)
		}
	})
}

func TestMaxBodyAndContentLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)

		sizeStr := r.URL.Query().Get("size")
		if sizeStr == "" {
			sizeStr = "100"
		}
		size, err := strconv.Atoi(sizeStr)
		if err != nil {
			size = 100
		}

		data := bytes.Repeat([]byte("a"), size)
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}))
	defer server.Close()

	t.Run("RequestBodyExceedsMax", func(t *testing.T) {
		body := bytes.Repeat([]byte("x"), 3000)
		opts := &RequestOptions{
			Method:        "POST",
			MaxBodyLength: 2000,
			Body:          body,
		}
		resp, err := Request("POST", server.URL, opts)
		if err == nil {
			t.Fatalf("Expected error due to exceeding MaxBodyLength, got success: %+v", resp)
		}
		t.Logf("RequestBodyExceedsMax: got error as expected: %v", err)
	})

	t.Run("RequestBodyWithinMax", func(t *testing.T) {
		body := bytes.Repeat([]byte("x"), 1000)
		opts := &RequestOptions{
			Method:        "POST",
			MaxBodyLength: 2000,
			Body:          body,
		}
		resp, err := Request("POST", server.URL, opts)
		if err != nil {
			t.Fatalf("Did not expect error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("ResponseExceedsMax", func(t *testing.T) {
		opts := &RequestOptions{
			MaxContentLength: 2000,
		}
		urlWithSize := server.URL + "?size=3000"
		resp, err := Get(urlWithSize, opts)
		if err == nil {
			t.Fatalf("Expected error due to exceeding MaxContentLength, got success: %+v", resp)
		}
		t.Logf("ResponseExceedsMax: got error as expected: %v", err)
	})

	t.Run("ResponseWithinMax", func(t *testing.T) {
		opts := &RequestOptions{
			MaxContentLength: 2000,
		}
		urlWithSize := server.URL + "?size=1000"
		resp, err := Get(urlWithSize, opts)
		if err != nil {
			t.Fatalf("Did not expect error, got %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", resp.StatusCode)
		}
		if len(resp.Body) != 1000 {
			t.Fatalf("Expected 1000 bytes, got %d", len(resp.Body))
		}
	})
}

func TestInterceptorErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "ok"}`))
	}))
	defer server.Close()

	t.Run("RequestInterceptorError", func(t *testing.T) {
		failingRequestInterceptor := func(req *http.Request) error {
			return fmt.Errorf("request interceptor forced error")
		}

		opts := &RequestOptions{
			InterceptorOptions: InterceptorOptions{
				RequestInterceptors: []func(*http.Request) error{
					failingRequestInterceptor,
				},
			},
		}

		resp, err := Get(server.URL, opts)
		if err == nil {
			t.Fatalf("Expected an error from request interceptor, got response: %+v", resp)
		}
		if !strings.Contains(err.Error(), "request interceptor forced error") {
			t.Errorf("Expected error to contain 'request interceptor forced error', got: %v", err)
		}
	})

	t.Run("ResponseInterceptorError", func(t *testing.T) {
		failingResponseInterceptor := func(resp *http.Response) error {
			return fmt.Errorf("response interceptor forced error")
		}

		opts := &RequestOptions{
			InterceptorOptions: InterceptorOptions{
				ResponseInterceptors: []func(*http.Response) error{
					failingResponseInterceptor,
				},
			},
		}

		resp, err := Get(server.URL, opts)
		if err == nil {
			t.Fatalf("Expected an error from response interceptor, got response: %+v", resp)
		}
		if !strings.Contains(err.Error(), "response interceptor forced error") {
			t.Errorf("Expected error to contain 'response interceptor forced error', got: %v", err)
		}
	})
}

func TestInvalidProxy(t *testing.T) {
	opts := &RequestOptions{
		Proxy: &Proxy{
			Protocol: "invalid-protocol",
			Host:     "bad_host",
			Port:     99999, // invalid port
		},
	}

	_, err := Get("http://example.com", opts)
	if err == nil {
		t.Fatal("Expected error due to invalid proxy settings, got nil")
	}

	t.Logf("InvalidProxy test got expected error: %v", err)
}

func TestInvalidMethod(t *testing.T) {
	opts := &RequestOptions{Method: "INVALID_METHOD!"}
	resp, err := Request("", "http://example.com", opts)
	if err == nil {
		t.Fatalf("Expected error for invalid method, got response: %v", resp)
	}
	if !strings.Contains(err.Error(), "invalid HTTP method") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestEmptyURL(t *testing.T) {
	opts := &RequestOptions{}

	resp, err := Get("", opts)
	if err == nil {
		t.Fatalf("Expected error for empty URL, but got response: %v", resp)
	}

	t.Logf("EmptyURL test got expected error: %v", err)
}

func TestNonHTTPBaseURL(t *testing.T) {
	client := NewClient("ftp://some.ftp.site")

	resp, err := client.Request(&RequestOptions{
		Method: "GET",
		URL:    "/test",
	})
	if err == nil {
		t.Fatalf("Expected error for non-HTTP base URL, got response: %v", resp)
	}

	t.Logf("NonHTTP_BaseURL test got expected error: %v", err)
}
