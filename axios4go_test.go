package axios4go

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get":
			json.NewEncoder(w).Encode(map[string]string{"message": "get success"})
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
