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
	"mime"
	"net/http"
	"strings"
)

// requireContentType reports whether the request has a Content-Type of
// text/plain; charset=utf-8. If it does not, a problem response is sent.
func requireContentType(w http.ResponseWriter, r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		sendProblem(
			w,
			r,
			http.StatusUnsupportedMediaType,
			problemMissingContentType,
			"Missing content type",
			"The Content-Type header must be specified and must be text/plain; charset=utf-8",
		)
		return false
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		sendProblem(
			w,
			r,
			http.StatusUnsupportedMediaType,
			problemBadContentType,
			"Invalid content type",
			err.Error(),
		)
		return false
	}

	if mediaType != "text/plain" || !strings.EqualFold(params["charset"], "utf-8") {
		sendProblem(
			w,
			r,
			http.StatusUnsupportedMediaType,
			problemUnsupportedContentType,
			"Unsupported content type",
			"The API requires the content be text/plain; charset=utf-8",
		)
		return false
	}

	return true
}

// requireAcceptJSON reports whether the request's Accept header allows an
// application/json response. A missing Accept header allows any type. If JSON
// is not acceptable, a problem response is sent.
//func requireAcceptJSON(w http.ResponseWriter, r *http.Request) bool {
//	header := r.Header.Get("Accept")
//	if strings.TrimSpace(header) == "" {
//		return true
//	}
//
//	accepted, malformed := acceptsJSON(header)
//	if accepted {
//		return true
//	}
//
//	if malformed {
//		sendProblem(
//			w,
//			r,
//			http.StatusNotAcceptable,
//			problemBadAcceptHeader,
//			"Accept header error",
//			"The Accept header could not be parsed",
//		)
//		return false
//	}
//
//	sendProblem(
//		w,
//		r,
//		http.StatusNotAcceptable,
//		problemUnsupportedResponseFormat,
//		"Unsupported response format",
//		"The API only supports application/json for the response format",
//	)
//	return false
//}

// acceptsJSON evaluates an Accept header against application/json. As in
// RFC 9110, the most specific matching range determines the quality value, so
// "application/json;q=0, */*" rejects JSON. Malformed ranges are skipped and
// reported through malformed so the caller can describe the failure.
//func acceptsJSON(header string) (accepted, malformed bool) {
//	bestSpecificity := -1
//	bestQuality := 0.0
//
//	for _, part := range strings.Split(header, ",") {
//		part = strings.TrimSpace(part)
//		if part == "" {
//			continue
//		}
//
//		mediaType, params, err := mime.ParseMediaType(part)
//		if err != nil {
//			malformed = true
//			continue
//		}
//
//		quality := 1.0
//		if q, ok := params["q"]; ok {
//			quality, err = strconv.ParseFloat(q, 64)
//			if err != nil || quality < 0 || quality > 1 {
//				malformed = true
//				continue
//			}
//		}
//
//		var specificity int
//		switch mediaType {
//		case "application/json":
//			specificity = 2
//		case "application/*":
//			specificity = 1
//		case "*/*":
//			specificity = 0
//		default:
//			continue
//		}
//
//		if specificity > bestSpecificity {
//			bestSpecificity = specificity
//			bestQuality = quality
//		}
//	}
//
//	return bestSpecificity >= 0 && bestQuality > 0, malformed
//}
