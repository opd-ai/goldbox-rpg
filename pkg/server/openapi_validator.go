package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/sirupsen/logrus"
)

// OpenAPIValidator validates HTTP requests and responses against OpenAPI specification
type OpenAPIValidator struct {
	router routers.Router
	spec   *openapi3.T
	mu     sync.RWMutex
}

// NewOpenAPIValidator creates a new validator from the OpenAPI spec file
func NewOpenAPIValidator(specPath string) (*OpenAPIValidator, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve spec path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("OpenAPI spec file not found: %s", absPath)
	}

	// Load OpenAPI spec
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	
	spec, err := loader.LoadFromFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the spec itself
	if err := spec.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	// Create router for matching paths
	router, err := legacy.NewRouter(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	return &OpenAPIValidator{
		router: router,
		spec:   spec,
	}, nil
}

// ValidateRequest validates an HTTP request against the OpenAPI spec
func (v *OpenAPIValidator) ValidateRequest(r *http.Request) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Find route matching the request
	route, pathParams, err := v.router.FindRoute(r)
	if err != nil {
		// Route not found in spec - skip validation
		return nil
	}

	// Create validation input
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}

	// Validate request
	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		return fmt.Errorf("request validation failed: %w", err)
	}

	return nil
}

// ValidateResponse validates an HTTP response against the OpenAPI spec
func (v *OpenAPIValidator) ValidateResponse(r *http.Request, resp *http.Response) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	// Find route matching the request
	route, pathParams, err := v.router.FindRoute(r)
	if err != nil {
		// Route not found in spec - skip validation
		return nil
	}

	// Create validation input
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
		},
		Status: resp.StatusCode,
		Header: resp.Header,
	}

	// Validate response (body validation requires reading it, which we skip for streaming)
	if err := openapi3filter.ValidateResponse(context.Background(), responseValidationInput); err != nil {
		return fmt.Errorf("response validation failed: %w", err)
	}

	return nil
}

// ValidationMiddleware returns HTTP middleware that validates requests
func (v *OpenAPIValidator) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := getLoggerFromContext(r.Context())

		// Validate request
		if err := v.ValidateRequest(r); err != nil {
			logger.WithFields(logrus.Fields{
				"method": r.Method,
				"path":   r.URL.Path,
				"error":  err.Error(),
			}).Warn("Request validation failed")

			// Return 400 Bad Request with validation error
			http.Error(w, fmt.Sprintf("Request validation failed: %v", err), http.StatusBadRequest)
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}
