package application_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/feature/user/application"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/mocks"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("GetUserProfileUseCase", func() {
	var (
		ctrl     *gomock.Controller
		mockRepo *mocks.MockUserRepository
		useCase  application.GetUserProfileUseCase
		ctx      context.Context
		testUser *entity.User
		userID   uuid.UUID
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockRepo = mocks.NewMockUserRepository(ctrl)
		useCase = application.NewGetUserProfileUseCase(mockRepo)
		ctx = context.Background()
		userID = uuid.New()
		testUser = entity.NewUser("firebase_test_uid", "firebase")
		testUser.ID = userID
		testUser.UpdateDisplayName("Test User")
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("Execute", func() {
		ginkgo.Context("when user exists", func() {
			ginkgo.It("should return user profile", func() {
				// Arrange
				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				// Act
				result, err := useCase.Execute(ctx, userID)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.ID).To(gomega.Equal(userID))
				gomega.Expect(result.FirebaseUID).To(gomega.Equal("firebase_test_uid"))
				gomega.Expect(*result.DisplayName).To(gomega.Equal("Test User"))
			})
		})

		ginkgo.Context("when user does not exist", func() {
			ginkgo.It("should return error", func() {
				// Arrange
				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(nil, entity.ErrUserNotFound)

				// Act
				result, err := useCase.Execute(ctx, userID)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrUserNotFound))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when user ID is nil", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := useCase.Execute(ctx, uuid.Nil)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidUserID))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when repository returns error", func() {
			ginkgo.It("should return repository error", func() {
				// Arrange
				expectedError := entity.ErrInvalidUserID
				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(nil, expectedError)

				// Act
				result, err := useCase.Execute(ctx, userID)

				// Assert
				gomega.Expect(err).To(gomega.Equal(expectedError))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})
