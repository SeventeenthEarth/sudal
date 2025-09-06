package integration_test

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	quizv1 "github.com/seventeenthearth/sudal/gen/go/quiz/v1"
	quizv1connect "github.com/seventeenthearth/sudal/gen/go/quiz/v1/quizv1connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/apispec"
	"github.com/seventeenthearth/sudal/internal/infrastructure/server"
	"github.com/seventeenthearth/sudal/internal/mocks"
	testhelpers "github.com/seventeenthearth/sudal/test/integration/helpers"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

// TDD: QuizService 통합 테스트 (현재는 실패가 정상)
// - 공개 RPC(gRPC-Web): CodeUnimplemented 기대
// - 보호 RPC(gRPC-Web): Unauthenticated 기대 (초기에는 적용 전이라 실패)
// - Connect(JSON): gRPC-only 필터로 404 기대 (초기에는 Quiz 경로 누락으로 실패)

var _ = Describe("QuizService gRPC integration (connect-go)", func() {
	var (
		ts      *testhelpers.TestServer
		baseURL string
	)

	BeforeEach(func() {
		mux := http.NewServeMux()

		// Build chains with mocked auth to exercise Unauthenticated on protected procedures
		logger := zap.NewNop()
		ctrl := gomock.NewController(GinkgoT())
		mockTV := mocks.NewMockTokenVerifier(ctrl)
		mockUS := mocks.NewMockUserService(ctrl)
		// Note: For missing Authorization header cases, interceptor returns Unauthenticated
		// without calling Verify or UserService; no expectations needed.
		mcb := server.NewMiddlewareChainBuilder(mockTV, mockUS, logger)
		chains := mcb.BuildServiceChains(apispec.ProtectedProcedures())

		// Use connect-go generated unimplemented stub for now
		unimpl := quizv1connect.UnimplementedQuizServiceHandler{}
		path, h := quizv1connect.NewQuizServiceHandler(unimpl, chains.SelectiveGRPC.ToConnectOptions()...)
		mux.Handle(path, chains.GRPCOnlyHTTP.Apply(h))

		var err error
		ts, err = testhelpers.NewTestServer(mux)
		Expect(err).NotTo(HaveOccurred())
		baseURL = ts.BaseURL
	})

	AfterEach(func() {
		if ts != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			Expect(ts.Close(ctx)).To(Succeed())
		}
	})

	Context("gRPC-Web public methods", func() {
		It("ListQuizSets should return CodeUnimplemented", func() {
			client := quizv1connect.NewQuizServiceClient(http.DefaultClient, baseURL, connect.WithGRPCWeb())
			req := connect.NewRequest(&quizv1.ListQuizSetsRequest{})
			resp, err := client.ListQuizSets(context.Background(), req)
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			ce, ok := err.(*connect.Error)
			Expect(ok).To(BeTrue())
			Expect(ce.Code()).To(Equal(connect.CodeUnimplemented))
		})

		It("GetQuizSet should return CodeUnimplemented", func() {
			client := quizv1connect.NewQuizServiceClient(http.DefaultClient, baseURL, connect.WithGRPCWeb())
			req := connect.NewRequest(&quizv1.GetQuizSetRequest{QuizSetId: 1})
			resp, err := client.GetQuizSet(context.Background(), req)
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			ce, ok := err.(*connect.Error)
			Expect(ok).To(BeTrue())
			Expect(ce.Code()).To(Equal(connect.CodeUnimplemented))
		})
	})

	Context("gRPC-Web protected methods (auth required)", func() {
		It("SubmitQuizResult should require authentication (Unauthenticated)", func() {
			client := quizv1connect.NewQuizServiceClient(http.DefaultClient, baseURL, connect.WithGRPCWeb())
			req := connect.NewRequest(&quizv1.SubmitQuizResultRequest{QuizSetId: 1})
			resp, err := client.SubmitQuizResult(context.Background(), req)
			// TDD: 현재는 인증 미적용/스텁이라 Unimplemented가 될 가능성 높음 -> 실패 유도
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			ce, ok := err.(*connect.Error)
			Expect(ok).To(BeTrue())
			Expect(ce.Code()).To(Equal(connect.CodeUnauthenticated))
		})

		It("GetUserQuizHistory should require authentication (Unauthenticated)", func() {
			client := quizv1connect.NewQuizServiceClient(http.DefaultClient, baseURL, connect.WithGRPCWeb())
			req := connect.NewRequest(&quizv1.GetUserQuizHistoryRequest{})
			resp, err := client.GetUserQuizHistory(context.Background(), req)
			// TDD: 현재는 인증 미적용/스텁이라 Unimplemented가 될 가능성 높음 -> 실패 유도
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			ce, ok := err.(*connect.Error)
			Expect(ok).To(BeTrue())
			Expect(ce.Code()).To(Equal(connect.CodeUnauthenticated))
		})
	})

	Context("Connect(JSON) should be blocked by gRPC-only filter", func() {
		It("ListQuizSets via Connect JSON returns 404", func() {
			// Raw HTTP request simulating Connect JSON
			url := baseURL + quizv1connect.QuizServiceListQuizSetsProcedure
			req, err := http.NewRequest("POST", url, http.NoBody)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Connect-Protocol-Version", "1")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = resp.Body.Close() }()

			// 현재는 Quiz 경로가 필터에 누락되어 있어 404가 아닐 가능성이 큼 -> 실패 유도
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
