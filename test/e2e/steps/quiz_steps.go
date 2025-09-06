package steps

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"strings"

	"connectrpc.com/connect"
	"github.com/cucumber/godog"

	quizv1 "github.com/seventeenthearth/sudal/gen/go/quiz/v1"
	quizv1connect "github.com/seventeenthearth/sudal/gen/go/quiz/v1/quizv1connect"
)

// QuizCtx holds the context for quiz-related test scenarios
type QuizCtx struct {
	baseURL    string
	httpClient *http.Client

	// Connect-go client (gRPC-Web or gRPC)
	grpcClient quizv1connect.QuizServiceClient

	// last RPC results
	lastRespList   *connect.Response[quizv1.ListQuizSetsResponse]
	lastRespGet    *connect.Response[quizv1.GetQuizSetResponse]
	lastRespSubmit *connect.Response[quizv1.SubmitQuizResultResponse]
	lastRespHist   *connect.Response[quizv1.GetUserQuizHistoryResponse]
	lastErr        error

	// raw HTTP result (for Connect JSON negative case)
	lastHTTPResp *http.Response
	lastHTTPBody []byte
}

func NewQuizCtx() *QuizCtx {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &QuizCtx{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (q *QuizCtx) Cleanup() {
	if q.lastHTTPResp != nil {
		_ = q.lastHTTPResp.Body.Close()
	}
}

func (q *QuizCtx) Register(sc *godog.ScenarioContext) {
	// We intentionally DO NOT register "the server is running" to avoid triple duplicates.
	sc.Step(`^the gRPC or gRPC-Web quiz client is connected$`, q.theGRPCOrGRPCWebClientIsConnected)

	sc.Step(`^I call ListQuizSets with empty request$`, q.iCallListQuizSetsEmpty)
	sc.Step(`^I call GetQuizSet with quiz_set_id "([^"]*)"$`, q.iCallGetQuizSetByID)
	sc.Step(`^I call SubmitQuizResult with quiz_set_id "([^"]*)"$`, q.iCallSubmitQuizResult)
	sc.Step(`^I call GetUserQuizHistory with empty request$`, q.iCallGetUserQuizHistoryEmpty)

	sc.Step(`^I have no authentication token$`, q.iHaveNoAuthToken)
	sc.Step(`^I should receive CodeUnimplemented error$`, q.iShouldReceiveUnimplemented)
	sc.Step(`^I should receive Unauthenticated error$`, q.iShouldReceiveUnauthenticated)
	sc.Step(`^the error message should contain "([^"]*)"$`, q.theErrorMessageShouldContain)

	sc.Step(`^I call ListQuizSets via Connect JSON protocol$`, q.iCallListQuizSetsViaConnectJSON)
	sc.Step(`^I should receive HTTP status (\d+)$`, q.iShouldReceiveHTTPStatus)
}

// Steps

func (q *QuizCtx) theGRPCOrGRPCWebClientIsConnected() error {
	// Prefer gRPC-Web for portability
	q.grpcClient = quizv1connect.NewQuizServiceClient(q.httpClient, q.baseURL, connect.WithGRPCWeb())
	return nil
}

func (q *QuizCtx) iCallListQuizSetsEmpty() error {
	if q.grpcClient == nil {
		return fmt.Errorf("grpc client not connected")
	}
	q.lastRespList, q.lastErr = q.grpcClient.ListQuizSets(context.Background(), connect.NewRequest(&quizv1.ListQuizSetsRequest{}))
	return nil
}

func (q *QuizCtx) iCallGetQuizSetByID(idStr string) error {
	if q.grpcClient == nil {
		return fmt.Errorf("grpc client not connected")
	}
	// parse ID best-effort (empty -> 0)
	id, _ := strconv.ParseInt(idStr, 10, 64)
	q.lastRespGet, q.lastErr = q.grpcClient.GetQuizSet(context.Background(), connect.NewRequest(&quizv1.GetQuizSetRequest{QuizSetId: id}))
	return nil
}

func (q *QuizCtx) iCallSubmitQuizResult(idStr string) error {
	if q.grpcClient == nil {
		return fmt.Errorf("grpc client not connected")
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	q.lastRespSubmit, q.lastErr = q.grpcClient.SubmitQuizResult(context.Background(), connect.NewRequest(&quizv1.SubmitQuizResultRequest{QuizSetId: id}))
	return nil
}

func (q *QuizCtx) iCallGetUserQuizHistoryEmpty() error {
	if q.grpcClient == nil {
		return fmt.Errorf("grpc client not connected")
	}
	q.lastRespHist, q.lastErr = q.grpcClient.GetUserQuizHistory(context.Background(), connect.NewRequest(&quizv1.GetUserQuizHistoryRequest{}))
	return nil
}

func (q *QuizCtx) iHaveNoAuthToken() error {
	// No-op: we simply don't set Authorization header in requests
	return nil
}

func (q *QuizCtx) iShouldReceiveUnimplemented() error {
	if q.lastErr == nil {
		return fmt.Errorf("expected error, got nil")
	}
	if ce, ok := q.lastErr.(*connect.Error); ok {
		if ce.Code() != connect.CodeUnimplemented {
			return fmt.Errorf("expected CodeUnimplemented, got %v", ce.Code())
		}
		return nil
	}
	return fmt.Errorf("expected *connect.Error, got %T", q.lastErr)
}

func (q *QuizCtx) iShouldReceiveUnauthenticated() error {
	if q.lastErr == nil {
		return fmt.Errorf("expected error, got nil")
	}
	if ce, ok := q.lastErr.(*connect.Error); ok {
		if ce.Code() != connect.CodeUnauthenticated {
			return fmt.Errorf("expected Unauthenticated, got %v", ce.Code())
		}
		return nil
	}
	return fmt.Errorf("expected *connect.Error, got %T", q.lastErr)
}

func (q *QuizCtx) theErrorMessageShouldContain(substr string) error {
	if q.lastErr == nil {
		return fmt.Errorf("no error to inspect")
	}
	if !strings.Contains(strings.ToLower(q.lastErr.Error()), strings.ToLower(substr)) {
		return fmt.Errorf("error message %q does not contain %q", q.lastErr.Error(), substr)
	}
	return nil
}

func (q *QuizCtx) iCallListQuizSetsViaConnectJSON() error {
	url := q.baseURL + quizv1connect.QuizServiceListQuizSetsProcedure
	req, err := http.NewRequest("POST", url, http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connect-Protocol-Version", "1")
	resp, err := q.httpClient.Do(req)
	q.lastHTTPResp = resp
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	body, readErr := io.ReadAll(resp.Body)
	q.lastHTTPBody = body
	if readErr != nil {
		return fmt.Errorf("failed to read response body: %w", readErr)
	}
	return nil
}

func (q *QuizCtx) iShouldReceiveHTTPStatus(status int) error {
	if q.lastHTTPResp == nil {
		return fmt.Errorf("no http response")
	}
	if q.lastHTTPResp.StatusCode != status {
		return fmt.Errorf("expected HTTP %d, got %d (body=%s)", status, q.lastHTTPResp.StatusCode, string(q.lastHTTPBody))
	}
	return nil
}

// no extra helpers required
