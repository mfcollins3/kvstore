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

package kvstore

// InMemoryStore implements a key-value store where the data is kept entirely in
// memory. Because the InMemoryStore keeps all data in memory, the data will be
// lost when the InMemoryStore goes away when the server program completes.
type InMemoryStore map[string]string

// NewInMemoryStore will create and return an empty InMemoryStore.
func NewInMemoryStore() InMemoryStore {
	return make(InMemoryStore)
}

// Get will search the InMemoryStore for a key-value pair indexed with the given
// key. If the key is found in the store, the value associated with the key will
// be returned.
//
// Get has two return values. If the key is found, the value will be returned
// in the value return value and the present value will be true to indicate that
// the key was found. If the key is not found, the value that is returned will
// be an empty string and the present value will be false.
func (s InMemoryStore) Get(key string) (value string, present bool) {
	value, present = s[key]
	return
}

// Set stores a key-value pair into the InMemoryStore. If the key already exists
// in the store, the value will be updated to the new value.
func (s InMemoryStore) Set(key, value string) {
	s[key] = value
}
