// Copyright 2026 Michael F. Collins, III
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package api_test

import (
	"encoding/json/v2"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/iotest"

	"go.michaelfcollins3.dev/kvstore/internal/api"
)

func TestSetValue(t *testing.T) {
	t.Run("the key does not exist", func(t *testing.T) {
		store := &mockStoreWithoutKey{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		request.SetPathValue("key", "hello")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it succeeded", func(t *testing.T) {
			if recorder.Code != http.StatusCreated {
				t.Errorf(
					"expected Created status, but received %d",
					recorder.Code,
				)
			}
		})

		t.Run("it stored the value in the store", func(t *testing.T) {
			store.Assert(t)
		})

		t.Run("it sets the Location header", func(t *testing.T) {
			if location := recorder.Header().Get("Location"); location != "/hello" {
				t.Errorf("expected Location to be %q, but got %q", "/hello", location)
			}
		})
	})

	t.Run("the key contains a slash", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/a%2Fb",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", "a/b")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if location := recorder.Header().Get("Location"); location != "/a%2Fb" {
			t.Errorf("expected Location to be %q, but got %q", "/a%2Fb", location)
		}
	})

	t.Run("the key is routed by a ServeMux", func(t *testing.T) {
		store := &storeStub{}
		mux := http.NewServeMux()
		mux.Handle("PUT /{key}", api.SetValue(store))
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/a%2Fb",
			strings.NewReader("world"),
		)
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")

		mux.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201 status code, but received %d", recorder.Code)
		}
	})

	t.Run("the key exists", func(t *testing.T) {
		store := &mockStoreWithKey{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("goodbye"),
		)
		request.SetPathValue("key", "hello")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it succeeded", func(t *testing.T) {
			if recorder.Code != http.StatusNoContent {
				t.Errorf("expected No Content status, but received %d", recorder.Code)
			}
		})

		t.Run("it updated the value in the store", func(t *testing.T) {
			store.Assert(t)
		})

		t.Run("it does not set the Location header", func(t *testing.T) {
			if location := recorder.Header().Get("Location"); location != "" {
				t.Errorf("expected no Location header, but got %q", location)
			}
		})
	})

	t.Run("the Content-Type header is not set", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", "hello")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it returns a 415 status code", func(t *testing.T) {
			if recorder.Code != http.StatusUnsupportedMediaType {
				t.Errorf(
					"expected 415 status code, but received %d",
					recorder.Code,
				)
			}
		})
	})

	t.Run("the wrong Content-Type is sent", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", "hello")
		request.Header.Set("Content-Type", "text/markdown")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it returns a 415 status code", func(t *testing.T) {
			if recorder.Code != http.StatusUnsupportedMediaType {
				t.Errorf(
					"expected 415 status code, but received %d",
					recorder.Code,
				)
			}
		})
	})

	t.Run("the request body times out", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			iotest.ErrReader(os.ErrDeadlineExceeded),
		)
		request.SetPathValue("key", "hello")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusRequestTimeout {
			t.Errorf(
				"expected 408 status code, but received %d",
				recorder.Code,
			)
		}

		assertProblemType(
			t,
			recorder,
			"http://go.michaelfcollins3.dev/kvstore/problems/request-timeout",
		)
	})

	t.Run(
		"an error occurs while reading the request body",
		func(t *testing.T) {
			store := &storeStub{}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				http.MethodPut,
				"/hello",
				iotest.ErrReader(errors.New("test error")),
			)
			request.SetPathValue("key", "hello")
			request.Header.Set("Content-Type", "text/plain; charset=utf-8")
			handler := api.SetValue(store)

			handler.ServeHTTP(recorder, request)

			t.Run("returns a 400 status code", func(t *testing.T) {
				if recorder.Code != http.StatusBadRequest {
					t.Errorf(
						"expected 400 status code, but received %d",
						recorder.Code,
					)
				}
			})
		},
	)

	t.Run("the request body is too large", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader(strings.Repeat("a", 1<<20+1)),
		)
		request.SetPathValue("key", "hello")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it returns a 413 status code", func(t *testing.T) {
			if recorder.Code != http.StatusRequestEntityTooLarge {
				t.Errorf(
					"expected 413 status code, but received %d",
					recorder.Code,
				)
			}
		})

		t.Run("it returns the correct error response", func(t *testing.T) {
			var response struct {
				Type   string `json:"type"`
				Title  string `json:"title,omitempty"`
				Detail string `json:"detail,omitempty"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			expected := "http://go.michaelfcollins3.dev/kvstore/problems/request-body-too-large"
			if response.Type != expected {
				t.Errorf("expected response type to be %q, but got %q", expected, response.Type)
			}
		})
	})

	t.Run("the request body is exactly the maximum size", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader(strings.Repeat("a", 1<<20)),
		)
		request.SetPathValue("key", "hello")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if recorder.Code == http.StatusRequestEntityTooLarge {
			t.Errorf("expected body at the limit to be accepted, but received 413")
		}
	})

	t.Run("bad content type", func(t *testing.T) {
		tests := []struct {
			name        string
			contentType string
		}{
			{"parameter without value", "text/plain; chatset"},
			{"missing subtype", "text/"},
			{"no slash and bad param", "text; =utf-8"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				store := &storeStub{}
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(
					http.MethodPut,
					"/hello",
					strings.NewReader("world"),
				)
				request.SetPathValue("key", "hello")
				request.Header.Set("Content-Type", tt.contentType)
				handler := api.SetValue(store)

				handler.ServeHTTP(recorder, request)

				if recorder.Code != http.StatusUnsupportedMediaType {
					t.Errorf(
						"expected 415 status code, but received %d",
						recorder.Code,
					)
				}
			})
		}
	})

	t.Run("the body is empty", func(t *testing.T) {
		store := &mockStoreWithoutKey{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader(""),
		)
		request.SetPathValue("key", "hello")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Errorf(
				"expected 201 status code, but received %d",
				recorder.Code,
			)
		}

		store.Assert(t)
	})

	t.Run("the key is too long", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", strings.Repeat("a", 1025))
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusRequestURITooLong {
			t.Errorf(
				"expected 414 status code, but received %d",
				recorder.Code,
			)
		}

		assertProblemType(
			t,
			recorder,
			"http://go.michaelfcollins3.dev/kvstore/problems/key-too-long",
		)
	})

	t.Run("the key is invalid and the Content-Type is missing", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", "hel\xfflo")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		assertProblemType(
			t,
			recorder,
			"http://go.michaelfcollins3.dev/kvstore/problems/invalid-utf8-key",
		)
	})

	t.Run("the key has the maximum length", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", strings.Repeat("a", 1024))
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Errorf(
				"expected 201 status code, but received %d",
				recorder.Code,
			)
		}
	})

	t.Run("the body has an invalid character", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("bad\xffvalue"),
		)
		request.SetPathValue("key", "hello")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf(
				"expected 400 status code, but received %d",
				recorder.Code,
			)
		}

		assertProblemType(
			t,
			recorder,
			"http://go.michaelfcollins3.dev/kvstore/problems/invalid-utf8-value",
		)
	})

	t.Run("the key is not valid UTF-8", func(t *testing.T) {
		store := &storeStub{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", "hel\xfflo")
		request.Header.Set("Content-Type", "text/plain; charset=utf-8")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf(
				"expected 400 status code, but received %d",
				recorder.Code,
			)
		}

		assertProblemType(
			t,
			recorder,
			"http://go.michaelfcollins3.dev/kvstore/problems/invalid-utf8-key",
		)
	})
}

func assertKeyValueBody(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	key, value string,
) {
	t.Helper()

	var response struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if response.Key != key || response.Value != value {
		t.Errorf(
			"expected key %q and value %q, but got key %q and value %q",
			key,
			value,
			response.Key,
			response.Value,
		)
	}
}

func assertProblemType(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	expected string,
) {
	t.Helper()

	var response struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if response.Type != expected {
		t.Errorf("expected response type to be %q, but got %q", expected, response.Type)
	}
}
