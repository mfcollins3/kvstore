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
	"strings"
	"testing"

	"go.michaelfcollins3.dev/kvstore/internal/api"
)

type mockStoreWithoutKey struct {
	calls int
}

func (s *mockStoreWithoutKey) Get(_ string) (string, bool) {
	return "", false
}

func (s *mockStoreWithoutKey) Set(key, _ string) {
	if key == "hello" {
		s.calls++
	}
}

func (s *mockStoreWithoutKey) Assert(t *testing.T) {
	if s.calls != 1 {
		t.Errorf("Expected Set to be called once, got %d", s.calls)
	}
}

type mockStoreWithKey struct {
	getCalls int
	setCalls int
}

func (s *mockStoreWithKey) Get(key string) (string, bool) {
	s.getCalls++
	return "test", true
}

func (s *mockStoreWithKey) Set(key, _ string) {
	if key == "hello" {
		s.setCalls++
	}
}

func (s *mockStoreWithKey) Assert(t *testing.T) {
	if s.getCalls != 1 {
		t.Errorf("Expected Get to be called once, got %d", s.getCalls)
	}

	if s.setCalls != 1 {
		t.Errorf("Expected Set to be called once, got %d", s.setCalls)
	}
}

func TestSetValue(t *testing.T) {
	t.Run("the key does not exist", func(t *testing.T) {
		store := &mockStoreWithoutKey{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPut,
			"/hello",
			strings.NewReader("world"),
		)
		request.SetPathValue("key", "hello")
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it succeeded", func(t *testing.T) {
			if recorder.Code != http.StatusOK {
				t.Errorf("expected OK status, but received %d", recorder.Code)
			}
		})

		t.Run("it stored the value in the store", func(t *testing.T) {
			store.Assert(t)
		})
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
		handler := api.SetValue(store)

		handler.ServeHTTP(recorder, request)

		t.Run("it succeeded", func(t *testing.T) {
			if recorder.Code != http.StatusOK {
				t.Errorf("expected OK status, but received %d", recorder.Code)
			}
		})

		t.Run("it updated the value in the store", func(t *testing.T) {
			store.Assert(t)
		})
	})
}
