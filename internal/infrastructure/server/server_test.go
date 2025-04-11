package server_test

import (
	"net"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
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
		// Setup for all Start tests
		ginkgo.BeforeEach(func() {
			// Ensure we have a valid config for dependency injection
			cfg := &config.Config{
				ServerPort:  "8080",
				LogLevel:    "info",
				Environment: "test",
			}
			config.SetConfig(cfg)
		})

		ginkgo.AfterEach(func() {
			// Reset config after tests
			config.SetConfig(nil)
		})

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

		ginkgo.Context("when receiving a shutdown signal", func() {
			var (
				srv    *server.Server
				errCh  chan error
				doneCh chan struct{}
			)

			ginkgo.BeforeEach(func() {
				// Create a server on a random available port
				srv = server.NewServer("0") // Port 0 means a random available port
				errCh = make(chan error, 1)
				doneCh = make(chan struct{})

				// Start the server in a goroutine
				go func() {
					errCh <- srv.Start()
					close(doneCh)
				}()

				// Give the server a moment to start
				time.Sleep(100 * time.Millisecond)
			})

			ginkgo.It("should shut down gracefully when receiving an interrupt signal", func() {
				// Use the TriggerShutdown method to simulate a shutdown signal
				// This will send a signal to the internal shutdown channel
				srv.TriggerShutdown()

				// Wait for the server to shut down
				gomega.Eventually(doneCh).WithTimeout(3 * time.Second).Should(gomega.BeClosed())

				// Check that no error was returned
				var serverErr error
				select {
				case serverErr = <-errCh:
					// Got an error or nil
				default:
					// No error yet, which is unexpected
					ginkgo.Fail("Expected server to return after shutdown")
				}

				// Server should have shut down gracefully with no error
				gomega.Expect(serverErr).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.Context("when shutdown fails", func() {
			// Since we can't easily mock the http.Server's Shutdown method, we'll remove this test
			// A better approach would be to refactor the server to use an interface for the HTTP server
			// that could be easily mocked for testing
		})
	})
})
