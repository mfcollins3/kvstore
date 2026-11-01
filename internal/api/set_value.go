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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"
)

const (
	// maxKeyBytes is the maximum length of a key, in bytes.
	maxKeyBytes = 1024

	// maxValueBytes is the maximum length of a value, in bytes.
	maxValueBytes = 1 << 20
)

// SetValue is a factory function that returns an HTTP handler to store a value
// into a key-value store.
//
// The HTTP request expected by SetValue specifies the key in the request path
// and contains the value in the body. SetValue is intended to be called using
// the PUT HTTP method.
//
// PUT /{key} HTTP/1.1
// Content-Type: text/plain; charset=utf-8
//
// {value}
//
// The key and the value must be valid UTF-8. The key must not exceed 1024
// bytes and the value must not exceed 1 MiB. The value may be empty. The
// request must have a Content-Type of text/plain; charset=utf-8, and the
// Accept header, if present, must allow application/json.
//
// The request is validated in the following order: the key, the Content-Type
// header, the Accept header, and then the body. The first failure is reported
// as an application/problem+json response.
//
// If the key did not exist in the store, SetValue responds with 201 Created and
// a Location header. If the key existed, the value is replaced and SetValue
// responds with 200 OK. Both responses contain a JSON object with the key and
// the value.
//
// SetValue is not safe for concurrent use unless the caller serializes access
// to the store, because the status code depends on a Get followed by a Set.
func SetValue(store keyValueStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if !validateKey(w, r, key) {
			return
		}

		if !requireContentType(w, r) {
			return
		}

		value, ok := readValue(w, r, key)
		if !ok {
			return
		}

		_, present := store.Get(key)
		store.Set(key, value)

		status := http.StatusNoContent
		if !present {
			w.Header().Set("Location", "/"+url.PathEscape(strings.ToLower(key)))
			status = http.StatusCreated
		}

		w.WriteHeader(status)
	})
}

// validateKey reports whether the key is acceptable. If it is not, a problem
// response is sent. The key itself is never logged because it is unvalidated.
func validateKey(w http.ResponseWriter, r *http.Request, key string) bool {
	if len(key) > maxKeyBytes {
		slog.DebugContext(r.Context(), "key too long", "length", len(key))
		sendProblem(
			w,
			r,
			http.StatusRequestURITooLong,
			problemKeyTooLong,
			"Key too long",
			fmt.Sprintf("The key must not exceed %d bytes", maxKeyBytes),
		)
		return false
	}

	if !utf8.ValidString(key) {
		slog.DebugContext(r.Context(), "invalid UTF-8 key", "length", len(key))
		sendProblem(
			w,
			r,
			http.StatusBadRequest,
			problemInvalidUTF8Key,
			"Invalid UTF-8 key",
			"The provided key is not valid UTF-8",
		)
		return false
	}

	return true
}

// readValue reads and validates the request body. If the body cannot be read or
// is not acceptable, a problem response is sent and ok is false.
func readValue(
	w http.ResponseWriter,
	r *http.Request,
	key string,
) (value string, ok bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxValueBytes))
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		switch {
		case errors.As(err, &maxBytesErr):
			slog.DebugContext(
				r.Context(),
				"request body too large",
				"key",
				key,
				"limit",
				maxBytesErr.Limit,
			)
			sendProblem(
				w,
				r,
				http.StatusRequestEntityTooLarge,
				problemBodyTooLarge,
				"Request body too large",
				fmt.Sprintf(
					"The request body must not exceed %d bytes",
					maxBytesErr.Limit,
				),
			)
		case isTimeout(err):
			slog.DebugContext(r.Context(), "timed out reading request body", "key", key)
			sendProblem(
				w,
				r,
				http.StatusRequestTimeout,
				problemRequestTimeout,
				"Request timeout",
				"The request body was not received in time.",
			)
		default:
			slog.ErrorContext(r.Context(), "failed to read request body", "error", err)
			sendProblem(
				w,
				r,
				http.StatusBadRequest,
				problemReadBody,
				"Failed to read request body",
				"The request body could not be read.",
			)
		}

		return "", false
	}

	if !utf8.Valid(body) {
		slog.DebugContext(r.Context(), "invalid UTF-8 value", "key", key)
		sendProblem(
			w,
			r,
			http.StatusBadRequest,
			problemInvalidUTF8Value,
			"Invalid UTF-8 value",
			"The provided value is not valid UTF-8",
		)
		return "", false
	}

	return string(body), true
}

// isTimeout reports whether err was caused by a deadline expiring.
func isTimeout(err error) bool {
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
