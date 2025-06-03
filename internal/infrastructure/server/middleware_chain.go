package server

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/firebase"
	"github.com/seventeenthearth/sudal/internal/infrastructure/middleware"
	"go.uber.org/zap"
)

// HTTPMiddleware represents an HTTP middleware function
type HTTPMiddleware func(http.Handler) http.Handler

// ConnectInterceptor represents a Connect-go interceptor
type ConnectInterceptor connect.UnaryInterceptorFunc

// MiddlewareChain represents a chain of HTTP middlewares
type MiddlewareChain struct {
	middlewares []HTTPMiddleware
}

// InterceptorChain represents a chain of Connect-go interceptors
type InterceptorChain struct {
	interceptors []ConnectInterceptor
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain(middlewares ...HTTPMiddleware) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: middlewares,
	}
}

// NewInterceptorChain creates a new interceptor chain
func NewInterceptorChain(interceptors ...ConnectInterceptor) *InterceptorChain {
	return &InterceptorChain{
		interceptors: interceptors,
	}
}

// Apply applies the middleware chain to an HTTP handler
func (mc *MiddlewareChain) Apply(handler http.Handler) http.Handler {
	// Apply middlewares in reverse order so they execute in the correct order
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		handler = mc.middlewares[i](handler)
	}
	return handler
}

// ToConnectOptions converts the interceptor chain to Connect-go options
func (ic *InterceptorChain) ToConnectOptions() []connect.HandlerOption {
	if len(ic.interceptors) == 0 {
		return nil
	}

	// Convert ConnectInterceptor to connect.Interceptor
	connectInterceptors := make([]connect.Interceptor, len(ic.interceptors))
	for i, interceptor := range ic.interceptors {
		connectInterceptors[i] = connect.UnaryInterceptorFunc(interceptor)
	}

	return []connect.HandlerOption{
		connect.WithInterceptors(connectInterceptors...),
	}
}

// MiddlewareChainBuilder helps build different types of middleware chains
type MiddlewareChainBuilder struct {
	firebaseHandler *firebase.FirebaseHandler
	logger          *zap.Logger
}

// NewMiddlewareChainBuilder creates a new middleware chain builder
func NewMiddlewareChainBuilder(firebaseHandler *firebase.FirebaseHandler, logger *zap.Logger) *MiddlewareChainBuilder {
	return &MiddlewareChainBuilder{
		firebaseHandler: firebaseHandler,
		logger:          logger,
	}
}

// PublicHTTPChain creates a middleware chain for public HTTP endpoints
func (mcb *MiddlewareChainBuilder) PublicHTTPChain() *MiddlewareChain {
	return NewMiddlewareChain(
		middleware.RequestLogger,
	)
}

// ProtectedHTTPChain creates a middleware chain for protected HTTP endpoints
func (mcb *MiddlewareChainBuilder) ProtectedHTTPChain() *MiddlewareChain {
	return NewMiddlewareChain(
		middleware.RequestLogger,
		middleware.AuthenticationMiddleware(mcb.firebaseHandler, mcb.logger),
	)
}

// GRPCOnlyHTTPChain creates a middleware chain for gRPC-only HTTP endpoints
func (mcb *MiddlewareChainBuilder) GRPCOnlyHTTPChain() *MiddlewareChain {
	return NewMiddlewareChain(
		middleware.ProtocolFilterMiddleware(middleware.GetGRPCOnlyPaths()),
		middleware.RequestLogger,
	)
}

// PublicGRPCChain creates an interceptor chain for public gRPC endpoints
func (mcb *MiddlewareChainBuilder) PublicGRPCChain() *InterceptorChain {
	return NewInterceptorChain()
}

// ProtectedGRPCChain creates an interceptor chain for protected gRPC endpoints
func (mcb *MiddlewareChainBuilder) ProtectedGRPCChain() *InterceptorChain {
	authInterceptor := ConnectInterceptor(middleware.AuthenticationInterceptor(mcb.firebaseHandler, mcb.logger))
	return NewInterceptorChain(authInterceptor)
}

// SelectiveGRPCChain creates an interceptor chain for selective gRPC authentication
func (mcb *MiddlewareChainBuilder) SelectiveGRPCChain(protectedProcedures []string) *InterceptorChain {
	authInterceptor := ConnectInterceptor(middleware.SelectiveAuthenticationInterceptor(mcb.firebaseHandler, mcb.logger, protectedProcedures))
	return NewInterceptorChain(authInterceptor)
}

// ServiceChains defines the middleware chains for different service types
type ServiceChains struct {
	// HTTP chains
	PublicHTTP    *MiddlewareChain
	ProtectedHTTP *MiddlewareChain
	GRPCOnlyHTTP  *MiddlewareChain

	// gRPC interceptor chains
	PublicGRPC    *InterceptorChain
	ProtectedGRPC *InterceptorChain
	SelectiveGRPC *InterceptorChain
}

// BuildServiceChains creates all the predefined service chains
func (mcb *MiddlewareChainBuilder) BuildServiceChains(protectedProcedures []string) *ServiceChains {
	return &ServiceChains{
		// HTTP chains
		PublicHTTP:    mcb.PublicHTTPChain(),
		ProtectedHTTP: mcb.ProtectedHTTPChain(),
		GRPCOnlyHTTP:  mcb.GRPCOnlyHTTPChain(),

		// gRPC interceptor chains
		PublicGRPC:    mcb.PublicGRPCChain(),
		ProtectedGRPC: mcb.ProtectedGRPCChain(),
		SelectiveGRPC: mcb.SelectiveGRPCChain(protectedProcedures),
	}
}
