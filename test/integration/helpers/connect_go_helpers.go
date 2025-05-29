package helpers

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"go.uber.org/mock/gomock"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

// ConnectGoMockHelper provides helper functions for configuring Connect-Go mock clients
type ConnectGoMockHelper struct {
	Mock     *mocks.MockConnectGoHealthServiceClient
	Protocol string
	Metadata map[string]string
}

// NewConnectGoMockHelper creates a new Connect-Go mock helper
func NewConnectGoMockHelper(ctrl *gomock.Controller, protocol string) *ConnectGoMockHelper {
	return &ConnectGoMockHelper{
		Mock:     mocks.NewMockConnectGoHealthServiceClient(ctrl),
		Protocol: protocol,
		Metadata: make(map[string]string),
	}
}

// SetServingStatus configures the mock to return SERVING status
func (h *ConnectGoMockHelper) SetServingStatus() {
	response := &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_SERVING,
	}

	h.Mock.EXPECT().
		Check(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *connect.Request[healthv1.CheckRequest]) (*connect.Response[healthv1.CheckResponse], error) {
			resp := connect.NewResponse(response)

			// Add protocol-specific headers
			switch h.Protocol {
			case "grpc-web":
				resp.Header().Set("Content-Type", "application/grpc-web+proto")
				resp.Header().Set("Grpc-Accept-Encoding", "gzip")
			case "http":
				resp.Header().Set("Content-Type", "application/json")
			default:
				resp.Header().Set("Content-Type", "application/grpc+proto")
			}

			// Add custom metadata
			for key, value := range h.Metadata {
				resp.Header().Set(key, value)
			}

			return resp, nil
		}).AnyTimes()
}

// SetNotServingStatus configures the mock to return NOT_SERVING status
func (h *ConnectGoMockHelper) SetNotServingStatus() {
	response := &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING,
	}

	h.Mock.EXPECT().
		Check(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *connect.Request[healthv1.CheckRequest]) (*connect.Response[healthv1.CheckResponse], error) {
			resp := connect.NewResponse(response)

			// Add protocol-specific headers
			switch h.Protocol {
			case "grpc-web":
				resp.Header().Set("Content-Type", "application/grpc-web+proto")
				resp.Header().Set("Grpc-Accept-Encoding", "gzip")
			case "http":
				resp.Header().Set("Content-Type", "application/json")
			default:
				resp.Header().Set("Content-Type", "application/grpc+proto")
			}

			// Add custom metadata
			for key, value := range h.Metadata {
				resp.Header().Set(key, value)
			}

			return resp, nil
		}).AnyTimes()
}

// SetUnknownStatus configures the mock to return UNKNOWN status
func (h *ConnectGoMockHelper) SetUnknownStatus() {
	response := &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED,
	}

	h.Mock.EXPECT().
		Check(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, req *connect.Request[healthv1.CheckRequest]) (*connect.Response[healthv1.CheckResponse], error) {
			resp := connect.NewResponse(response)

			// Add protocol-specific headers
			switch h.Protocol {
			case "grpc-web":
				resp.Header().Set("Content-Type", "application/grpc-web+proto")
				resp.Header().Set("Grpc-Accept-Encoding", "gzip")
			case "http":
				resp.Header().Set("Content-Type", "application/json")
			default:
				resp.Header().Set("Content-Type", "application/grpc+proto")
			}

			// Add custom metadata
			for key, value := range h.Metadata {
				resp.Header().Set(key, value)
			}

			return resp, nil
		}).AnyTimes()
}

// SetError configures the mock to return an error
func (h *ConnectGoMockHelper) SetError(err error) {
	h.Mock.EXPECT().
		Check(gomock.Any(), gomock.Any()).
		Return(nil, err).
		AnyTimes()
}

// SetConnectError configures the mock to return a Connect error
func (h *ConnectGoMockHelper) SetConnectError(code connect.Code, message string) {
	err := connect.NewError(code, fmt.Errorf("%s", message))
	h.SetError(err)
}

// AddMetadata adds custom metadata to responses
func (h *ConnectGoMockHelper) AddMetadata(key, value string) {
	h.Metadata[key] = value
}

// GetProtocol returns the protocol this mock client is configured for
func (h *ConnectGoMockHelper) GetProtocol() string {
	return h.Protocol
}

// GetMock returns the underlying mock client
func (h *ConnectGoMockHelper) GetMock() *mocks.MockConnectGoHealthServiceClient {
	return h.Mock
}

// GRPCMockHelper provides helper functions for configuring gRPC mock clients
type GRPCMockHelper struct {
	Mock     *mocks.MockGRPCHealthServiceClient
	Metadata map[string]string
}

// NewGRPCMockHelper creates a new gRPC mock helper
func NewGRPCMockHelper(ctrl *gomock.Controller) *GRPCMockHelper {
	return &GRPCMockHelper{
		Mock:     mocks.NewMockGRPCHealthServiceClient(ctrl),
		Metadata: make(map[string]string),
	}
}

// SetServingStatus configures the mock to return SERVING status
func (h *GRPCMockHelper) SetServingStatus() {
	response := &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_SERVING,
	}

	h.Mock.EXPECT().
		Check(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(response, nil).
		AnyTimes()
}

// SetNotServingStatus configures the mock to return NOT_SERVING status
func (h *GRPCMockHelper) SetNotServingStatus() {
	response := &healthv1.CheckResponse{
		Status: healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING,
	}

	h.Mock.EXPECT().
		Check(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(response, nil).
		AnyTimes()
}

// SetError configures the mock to return an error
func (h *GRPCMockHelper) SetError(err error) {
	h.Mock.EXPECT().
		Check(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, err).
		AnyTimes()
}

// GetMock returns the underlying mock client
func (h *GRPCMockHelper) GetMock() *mocks.MockGRPCHealthServiceClient {
	return h.Mock
}

// Legacy compatibility will be handled by updating the existing functions in grpc_client_mocks.go
