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
	"net/http"
	"net/http/httptest"
	"testing"

	"go.michaelfcollins3.dev/kvstore/internal/api"
)

type storeStub struct {
	HasValue bool
}

func (s *storeStub) Get(key string) (string, bool) {
	if key == "hello" {
		return "world", s.HasValue
	}

	return "", false
}

func (s *storeStub) Set(_, _ string) {
}

func TestGetValue(t *testing.T) {
	t.Run("the key exists", func(t *testing.T) {
		store := &storeStub{HasValue: true}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/hello", nil)
		request.SetPathValue("key", "hello")
		handler := api.GetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it returns a 200 status code", func(t *testing.T) {
			if recorder.Code != http.StatusOK {
				t.Errorf(
					"Expected a 200 status code, but got %d",
					recorder.Code,
				)
			}
		})

		t.Run("it returns the value in the body", func(t *testing.T) {
			value := recorder.Body.String()
			if value != "world" {
				t.Errorf("The value should be 'world', but got '%s'", value)
			}
		})
	})

	t.Run("the key does not exist", func(t *testing.T) {
		store := &storeStub{HasValue: false}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/hello", nil)
		request.SetPathValue("key", "hello")
		handler := api.GetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it returns a 404 status code", func(t *testing.T) {
			if recorder.Code != http.StatusNotFound {
				t.Errorf(
					"Expected a 404 status code, but got %d",
					recorder.Code,
				)
			}
		})
	})
}
