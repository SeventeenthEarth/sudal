package server_test

import (
	"net"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/infrastructure/server"
)

var _ = ginkgo.Describe("Server", func() {
	ginkgo.Describe("NewServer", func() {
		ginkgo.Context("when creating a server with empty port", func() {
			ginkgo.It("should return a non-nil server", func() {
				// Act
				srv := server.NewServer("")

				// Assert
				gomega.Expect(srv).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("when creating a server with specific port", func() {
			ginkgo.It("should return a non-nil server", func() {
				// Act
				srv := server.NewServer("9090")

				// Assert
				gomega.Expect(srv).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Start", func() {
		ginkgo.Context("when the port is already in use", func() {
			var (
				listener net.Listener
				srv      *server.Server
				errCh    chan error
			)

			ginkgo.BeforeEach(func() {
				// Arrange - create a server with a port that's already in use
				// First, start a server on port 9092
				var err error
				listener, err = net.Listen("tcp", ":9092")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Now create our test server on the same port
				srv = server.NewServer("9092")
				errCh = make(chan error, 1)

				// Start the server and expect an error
				go func() {
					errCh <- srv.Start()
				}()
			})

			ginkgo.AfterEach(func() {
				if listener != nil {
					_ = listener.Close()
				}
			})

			ginkgo.It("should return an error", func() {
				// Wait for the error
				var err error
				gomega.Eventually(func() error {
					select {
					case err = <-errCh:
						return err
					default:
						return nil
					}
				}).WithTimeout(2 * time.Second).ShouldNot(gomega.BeNil())
			})
		})
	})
})
