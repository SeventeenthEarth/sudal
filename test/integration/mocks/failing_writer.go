package mocks

import (
	"fmt"
	"net/http"
)

// FailingResponseWriter is a mock ResponseWriter that fails on Write
type FailingResponseWriter struct {
	Headers http.Header
	Code    int
	Body    []byte
	Err     error
}

// NewFailingResponseWriter creates a new FailingResponseWriter
func NewFailingResponseWriter() *FailingResponseWriter {
	return &FailingResponseWriter{
		Headers: make(http.Header),
		Err:     fmt.Errorf("write error"),
	}
}

// Header returns the response headers
func (f *FailingResponseWriter) Header() http.Header {
	return f.Headers
}

// WriteHeader sets the response status code
func (f *FailingResponseWriter) WriteHeader(code int) {
	f.Code = code
}

// Write fails with the configured error
func (f *FailingResponseWriter) Write(b []byte) (int, error) {
	f.Body = b // Store the body for inspection
	return 0, f.Err
}
