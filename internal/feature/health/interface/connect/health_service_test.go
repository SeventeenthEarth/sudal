package connect_test

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	connectImpl "github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = Describe("HealthServiceHandler", func() {
	var (
		mockCtrl      *gomock.Controller
		mockService   *mocks.MockService
		healthHandler *connectImpl.HealthServiceHandler
		ctx           context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockService = mocks.NewMockService(mockCtrl)
		healthHandler = connectImpl.NewHealthServiceHandler(mockService)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Check", func() {
		Context("when the service returns a healthy status", func() {
			BeforeEach(func() {
				mockService.EXPECT().
					Check(gomock.Any()).
					Return(domain.HealthyStatus(), nil)
			})

			It("should return a SERVING status", func() {
				// Arrange
				req := connect.NewRequest(&healthv1.CheckRequest{})

				// Act
				resp, err := healthHandler.Check(ctx, req)

				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))
			})
		})

		Context("when the service returns an unhealthy status", func() {
			BeforeEach(func() {
				mockService.EXPECT().
					Check(gomock.Any()).
					Return(domain.NewStatus("unhealthy"), nil)
			})

			It("should return a NOT_SERVING status", func() {
				// Arrange
				req := connect.NewRequest(&healthv1.CheckRequest{})

				// Act
				resp, err := healthHandler.Check(ctx, req)

				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING))
			})
		})

		Context("when the service returns an unknown status", func() {
			BeforeEach(func() {
				mockService.EXPECT().
					Check(gomock.Any()).
					Return(domain.NewStatus("unknown_status"), nil)
			})

			It("should return an UNKNOWN status", func() {
				// Arrange
				req := connect.NewRequest(&healthv1.CheckRequest{})

				// Act
				resp, err := healthHandler.Check(ctx, req)

				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED))
			})
		})

		Context("when the service returns an error", func() {
			BeforeEach(func() {
				mockService.EXPECT().
					Check(gomock.Any()).
					Return(nil, errors.New("service error"))
			})

			It("should return a connect error with internal code", func() {
				// Arrange
				req := connect.NewRequest(&healthv1.CheckRequest{})

				// Act
				resp, err := healthHandler.Check(ctx, req)

				// Assert
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())

				// Check that it's a connect error with the correct code
				connectErr, ok := err.(*connect.Error)
				Expect(ok).To(BeTrue())
				Expect(connectErr.Code()).To(Equal(connect.CodeInternal))
			})
		})
	})
})
