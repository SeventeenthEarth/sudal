package protocol_test

import (
	"context"

	"connectrpc.com/connect"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	quizv1 "github.com/seventeenthearth/sudal/gen/go/quiz/v1"
	quizv1connect "github.com/seventeenthearth/sudal/gen/go/quiz/v1/quizv1connect"
	protocol "github.com/seventeenthearth/sudal/internal/feature/quiz/protocol"
)

var _ = ginkgo.Describe("QuizManager (protocol)", func() {
	var svc *protocol.QuizManager

	ginkgo.BeforeEach(func() {
		svc = protocol.NewQuizManager(nil)
	})

	ginkgo.It("should be constructible", func() {
		// Returns a non-nil instance; nil logger is allowed (uses Nop).
		gomega.Expect(svc).NotTo(gomega.BeNil())
	})

	ginkgo.Describe("unimplemented behavior", func() {
		ginkgo.It("ListQuizSets returns CodeUnimplemented with message", func() {
			req := connect.NewRequest(&quizv1.ListQuizSetsRequest{})
			resp, err := svc.ListQuizSets(context.Background(), req)
			gomega.Expect(resp).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.HaveOccurred())
			if ce, ok := err.(*connect.Error); ok {
				gomega.Expect(ce.Code()).To(gomega.Equal(connect.CodeUnimplemented))
				gomega.Expect(ce.Message()).To(gomega.ContainSubstring("not implemented"))
			} else {
				gomega.Expect(ok).To(gomega.BeTrue(), "error should be *connect.Error")
			}
		})

		ginkgo.It("GetQuizSet returns CodeUnimplemented with message", func() {
			req := connect.NewRequest(&quizv1.GetQuizSetRequest{QuizSetId: 1})
			resp, err := svc.GetQuizSet(context.Background(), req)
			gomega.Expect(resp).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.HaveOccurred())
			if ce, ok := err.(*connect.Error); ok {
				gomega.Expect(ce.Code()).To(gomega.Equal(connect.CodeUnimplemented))
				gomega.Expect(ce.Message()).To(gomega.ContainSubstring("not implemented"))
			} else {
				gomega.Expect(ok).To(gomega.BeTrue(), "error should be *connect.Error")
			}
		})

		ginkgo.It("SubmitQuizResult returns CodeUnimplemented", func() {
			// Currently Unimplemented at handler level; auth is enforced in middleware.
			req := connect.NewRequest(&quizv1.SubmitQuizResultRequest{QuizSetId: 1})
			resp, err := svc.SubmitQuizResult(context.Background(), req)
			gomega.Expect(resp).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.HaveOccurred())
			if ce, ok := err.(*connect.Error); ok {
				gomega.Expect(ce.Code()).To(gomega.Equal(connect.CodeUnimplemented))
			} else {
				gomega.Expect(ok).To(gomega.BeTrue(), "error should be *connect.Error")
			}
		})

		ginkgo.It("GetUserQuizHistory returns CodeUnimplemented", func() {
			req := connect.NewRequest(&quizv1.GetUserQuizHistoryRequest{})
			resp, err := svc.GetUserQuizHistory(context.Background(), req)
			gomega.Expect(resp).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.HaveOccurred())
			if ce, ok := err.(*connect.Error); ok {
				gomega.Expect(ce.Code()).To(gomega.Equal(connect.CodeUnimplemented))
			} else {
				gomega.Expect(ok).To(gomega.BeTrue(), "error should be *connect.Error")
			}
		})

		ginkgo.It("should accept context (propagation sanity)", func() {
			// Ensure passing a context does not cause panics; functional logic is tested elsewhere.
			ctx := context.WithValue(context.Background(), struct{}{}, "x")
			_, err := svc.ListQuizSets(ctx, connect.NewRequest(&quizv1.ListQuizSetsRequest{}))
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})

// Interface compliance: ensure QuizManager implements the service handler.
var _ quizv1connect.QuizServiceHandler = (*protocol.QuizManager)(nil)
