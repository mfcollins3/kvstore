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
	"log/slog"
	"net/http"
)

// Problem types returned by the API.
const (
	problemBase = "http://go.michaelfcollins3.dev/kvstore/problems/"

	problemMissingContentType     = problemBase + "missing-content-type"
	problemBadContentType         = problemBase + "bad-content-type"
	problemUnsupportedContentType = problemBase + "unsupported-content-type"
	problemKeyTooLong             = problemBase + "key-too-long"
	problemInvalidUTF8Key         = problemBase + "invalid-utf8-key"
	problemInvalidUTF8Value       = problemBase + "invalid-utf8-value"
	problemBodyTooLarge           = problemBase + "request-body-too-large"
	problemRequestTimeout         = problemBase + "request-timeout"
	problemReadBody               = problemBase + "read-request-body"
)

type problemResponse struct {
	Type   string `json:"type"`
	Title  string `json:"title,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// sendProblem sends an application/problem+json response built from the given
// problem type, title and detail.
func sendProblem(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	problemType, title, detail string,
) {
	sendProblemResponse(w, r, statusCode, problemResponse{
		Type:   problemType,
		Title:  title,
		Detail: detail,
	})
}

func sendProblemResponse(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	response problemResponse,
) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to marshal problem response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(responseBytes)
}
