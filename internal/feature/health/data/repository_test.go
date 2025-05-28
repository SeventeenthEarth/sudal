package data_test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/feature/health/data"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

var _ = ginkgo.Describe("Repository", func() {
	ginkgo.Describe("NewRepository", func() {
		ginkgo.It("should create a new repository", func() {
			// Act
			repo := data.NewRepository(nil) // nil for test environment

			// Assert
			gomega.Expect(repo).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("GetStatus", func() {
		var (
			repo   *data.Repository
			ctx    context.Context
			status *domain.Status
			err    error
		)

		ginkgo.BeforeEach(func() {
			repo = data.NewRepository(nil) // nil for test environment
			ctx = context.Background()
		})

		ginkgo.JustBeforeEach(func() {
			status, err = repo.GetStatus(ctx)
		})

		ginkgo.It("should return a healthy status without error", func() {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal("healthy"))
		})
	})

	ginkgo.Describe("GetDatabaseStatus", func() {
		var (
			repo           *data.Repository
			ctx            context.Context
			databaseStatus *domain.DatabaseStatus
			err            error
		)

		ginkgo.BeforeEach(func() {
			repo = data.NewRepository(nil) // nil for test environment
			ctx = context.Background()
		})

		ginkgo.JustBeforeEach(func() {
			databaseStatus, err = repo.GetDatabaseStatus(ctx)
		})

		ginkgo.It("should return a healthy database status without error", func() {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(databaseStatus).NotTo(gomega.BeNil())
			gomega.Expect(databaseStatus.Status).To(gomega.Equal("healthy"))
			gomega.Expect(databaseStatus.Message).To(gomega.Equal("Mock database connection is healthy"))
			gomega.Expect(databaseStatus.Stats).NotTo(gomega.BeNil())
			gomega.Expect(databaseStatus.Stats.MaxOpenConnections).To(gomega.Equal(25))
		})
	})
})
