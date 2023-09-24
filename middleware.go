package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Config struct {
	// Define your configuration options here.
}

type CustomMetricsMiddleware struct {
	next      http.Handler
	config    *Config
	endpoints map[string]*EndpointMetrics
	mu        sync.Mutex
}

type EndpointMetrics struct {
	RequestCount    int
	ErrorResponse   int
	SuccessResponse int
	DurationSum     time.Duration
}

func New(next http.Handler, config *Config) (*CustomMetricsMiddleware, error) {
	return &CustomMetricsMiddleware{
		next:      next,
		config:    config,
		endpoints: make(map[string]*EndpointMetrics),
	}, nil
}

func (m *CustomMetricsMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	startTime := time.Now()

	// Create a custom response writer to capture the status code.
	customResponseWriter := NewCustomResponseWriter(rw)

	// Call the next middleware or handler in the chain.
	m.next.ServeHTTP(customResponseWriter, req)

	// Calculate duration.
	duration := time.Since(startTime)

	// Extract the endpoint path without query parameters.
	endpointPath := removeQueryParams(req.URL.Path)

	// Generate an endpoint key.
	endpointKey := fmt.Sprintf("%s_%s", req.Method, endpointPath)

	// Update metrics for the specific endpoint.
	m.mu.Lock()
	defer m.mu.Unlock()
	if metrics, ok := m.endpoints[endpointKey]; ok {
		metrics.RequestCount++
		if customResponseWriter.Status() >= 400 {
			metrics.ErrorResponse++
		} else {
			metrics.SuccessResponse++
		}
		metrics.DurationSum += duration
	} else {
		m.endpoints[endpointKey] = &EndpointMetrics{
			RequestCount:    1,
			ErrorResponse:   0,
			SuccessResponse: 0,
			DurationSum:     duration,
		}
	}
}

func CreateConfig() interface{} {
	return &Config{}
}

func GetConfig() interface{} {
	return &Config{}
}

func GetMiddleware() (http.Handler, error) {
	return &CustomMetricsMiddleware{}, nil
}

func removeQueryParams(path string) string {
	// Split the path by "?" to separate the path from the query parameters.
	parts := strings.Split(path, "?")
	if len(parts) > 0 {
		return parts[0] // Return only the path without query parameters.
	}
	return path
}

// CustomResponseWriter is a custom implementation of http.ResponseWriter
type CustomResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewCustomResponseWriter(rw http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		ResponseWriter: rw,
		status:         http.StatusOK,
	}
}

func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	crw.status = statusCode
	crw.ResponseWriter.WriteHeader(statusCode)
}

func (crw *CustomResponseWriter) Status() int {
	return crw.status
}
