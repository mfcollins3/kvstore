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

package api

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type trackingBody struct {
	reader *strings.Reader
	closed bool
}

func (b *trackingBody) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

func TestSendProblemResponse(t *testing.T) {
	t.Run("a complete problem response is sent", func(t *testing.T) {
		body := &trackingBody{reader: strings.NewReader("request body")}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/hello", body)
		problem := problemResponse{
			Type:   "http://go.michaelfcollins3.dev/kvstore/problems/test-problem",
			Title:  "Test Problem",
			Detail: "This is a test problem.",
		}

		sendProblemResponse(recorder, request, http.StatusBadRequest, problem)

		t.Run("it returns the status code", func(t *testing.T) {
			if recorder.Code != http.StatusBadRequest {
				t.Errorf(
					"expected 400 status code, but received %d",
					recorder.Code,
				)
			}
		})

		t.Run("it sets the Content-Type header", func(t *testing.T) {
			contentType := recorder.Header().Get("Content-Type")
			if contentType != "application/problem+json" {
				t.Errorf(
					"expected Content-Type to be %q, but got %q",
					"application/problem+json",
					contentType,
				)
			}
		})

		t.Run("it returns the problem in the response body", func(t *testing.T) {
			var response problemResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if response != problem {
				t.Errorf("expected response to be %+v, but got %+v", problem, response)
			}
		})

	})

	t.Run("only the problem type is set", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/hello", http.NoBody)
		problem := problemResponse{
			Type: "http://go.michaelfcollins3.dev/kvstore/problems/test-problem",
		}

		sendProblemResponse(recorder, request, http.StatusNotFound, problem)

		t.Run("it returns the status code", func(t *testing.T) {
			if recorder.Code != http.StatusNotFound {
				t.Errorf(
					"expected 404 status code, but received %d",
					recorder.Code,
				)
			}
		})

		t.Run("it omits the empty fields from the response body", func(t *testing.T) {
			var response map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if response["type"] != problem.Type {
				t.Errorf(
					"expected response type to be %q, but got %q",
					problem.Type,
					response["type"],
				)
			}

			for _, field := range []string{"title", "detail"} {
				if _, ok := response[field]; ok {
					t.Errorf("expected %q to be omitted from the response body", field)
				}
			}
		})
	})

	t.Run("the problem cannot be marshaled", func(t *testing.T) {
		body := &trackingBody{reader: strings.NewReader("request body")}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPut, "/hello", body)
		problem := problemResponse{
			// encoding/json/v2 rejects strings that are not valid UTF-8.
			Type: "http://go.michaelfcollins3.dev/kvstore/problems/\xff",
		}

		sendProblemResponse(recorder, request, http.StatusBadRequest, problem)

		t.Run("it returns a 500 status code", func(t *testing.T) {
			if recorder.Code != http.StatusInternalServerError {
				t.Errorf(
					"expected 500 status code, but received %d",
					recorder.Code,
				)
			}
		})

		t.Run("it does not set the Content-Type header", func(t *testing.T) {
			contentType := recorder.Header().Get("Content-Type")
			if contentType == "application/problem+json" {
				t.Errorf(
					"expected Content-Type not to be %q",
					"application/problem+json",
				)
			}
		})

		t.Run("it does not write a response body", func(t *testing.T) {
			if recorder.Body.Len() != 0 {
				t.Errorf(
					"expected an empty response body, but got %q",
					recorder.Body.String(),
				)
			}
		})
	})
}
