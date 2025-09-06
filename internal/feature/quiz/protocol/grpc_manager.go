package protocol

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	quizv1 "github.com/seventeenthearth/sudal/gen/go/quiz/v1"
	"github.com/seventeenthearth/sudal/gen/go/quiz/v1/quizv1connect"
	"go.uber.org/zap"
)

// QuizManager implements the Connect-go QuizService
// Initial scaffolding: methods return CodeUnimplemented to indicate not implemented yet.
type QuizManager struct {
	quizv1connect.UnimplementedQuizServiceHandler
	logger *zap.Logger
}

// shared unimplemented error to avoid repeated allocations
var errUnimplemented = connect.NewError(connect.CodeUnimplemented, errors.New("not implemented"))

// NewQuizManager creates a new QuizManager instance.
func NewQuizManager(logger *zap.Logger) *QuizManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &QuizManager{logger: logger}
}

// ListQuizSets returns unimplemented during scaffolding phase.
func (q *QuizManager) ListQuizSets(ctx context.Context, req *connect.Request[quizv1.ListQuizSetsRequest]) (*connect.Response[quizv1.ListQuizSetsResponse], error) {
	q.logger.Info("ListQuizSets called (scaffold)")
	return nil, errUnimplemented
}

// GetQuizSet returns unimplemented during scaffolding phase.
func (q *QuizManager) GetQuizSet(ctx context.Context, req *connect.Request[quizv1.GetQuizSetRequest]) (*connect.Response[quizv1.GetQuizSetResponse], error) {
	q.logger.Info("GetQuizSet called (scaffold)", zap.Int64("quiz_set_id", req.Msg.GetQuizSetId()))
	return nil, errUnimplemented
}

// SubmitQuizResult returns unimplemented during scaffolding phase.
// Authentication is enforced via SelectiveAuthenticationInterceptor in the chain.
func (q *QuizManager) SubmitQuizResult(ctx context.Context, req *connect.Request[quizv1.SubmitQuizResultRequest]) (*connect.Response[quizv1.SubmitQuizResultResponse], error) {
	q.logger.Info("SubmitQuizResult called (scaffold)", zap.Int64("quiz_set_id", req.Msg.GetQuizSetId()))
	return nil, errUnimplemented
}

// GetUserQuizHistory returns unimplemented during scaffolding phase.
// Authentication is enforced via SelectiveAuthenticationInterceptor in the chain.
func (q *QuizManager) GetUserQuizHistory(ctx context.Context, req *connect.Request[quizv1.GetUserQuizHistoryRequest]) (*connect.Response[quizv1.GetUserQuizHistoryResponse], error) {
	q.logger.Info("GetUserQuizHistory called (scaffold)")
	return nil, errUnimplemented
}
