package di_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
)

var _ = ginkgo.Describe("DI", func() {
	ginkgo.Describe("InitializeHealthHandler", func() {
		ginkgo.It("should return a non-nil handler", func() {
			// Act
			handler := di.InitializeHealthHandler()

			// Assert
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("InitializeHealthConnectHandler", func() {
		ginkgo.It("should return a non-nil Connect handler", func() {
			// Act
			handler := di.InitializeHealthConnectHandler()

			// Assert
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("ProvideConfig", func() {
		ginkgo.Context("when config is loaded", func() {
			var cfg *config.Config

			ginkgo.BeforeEach(func() {
				// Load and set a config
				var err error
				cfg, err = config.LoadConfig("")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				config.SetConfig(cfg)
			})

			ginkgo.It("should return the config instance", func() {
				// Act
				providedCfg := di.ProvideConfig()

				// Assert
				gomega.Expect(providedCfg).To(gomega.Equal(cfg))
				gomega.Expect(providedCfg).To(gomega.Equal(config.GetConfig()))
			})
		})

		ginkgo.Context("when config is not loaded", func() {
			ginkgo.BeforeEach(func() {
				// Reset the config instance
				config.SetConfig(nil)
			})

			ginkgo.It("should panic when trying to provide config", func() {
				// Act & Assert
				gomega.Expect(func() {
					di.ProvideConfig()
				}).To(gomega.Panic())
			})
		})
	})
})
