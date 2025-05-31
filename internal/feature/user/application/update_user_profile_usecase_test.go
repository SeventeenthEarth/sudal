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

var _ = ginkgo.Describe("UpdateUserProfileUseCase", func() {
	var (
		ctrl     *gomock.Controller
		mockRepo *mocks.MockUserRepository
		useCase  application.UpdateUserProfileUseCase
		ctx      context.Context
		testUser *entity.User
		userID   uuid.UUID
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockRepo = mocks.NewMockUserRepository(ctrl)
		useCase = application.NewUpdateUserProfileUseCase(mockRepo)
		ctx = context.Background()
		userID = uuid.New()
		testUser = entity.NewUser("firebase_test_uid", "firebase")
		testUser.ID = userID
		testUser.UpdateDisplayName("Original Name")
		testUser.UpdateAvatarURL("https://example.com/original.jpg")
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("Execute", func() {
		ginkgo.Context("when updating display name and avatar URL", func() {
			ginkgo.It("should update both fields successfully", func() {
				// Arrange
				newDisplayName := "Updated Name"
				newAvatarURL := "https://example.com/updated.jpg"

				updatedUser := *testUser
				updatedUser.UpdateDisplayName(newDisplayName)
				updatedUser.UpdateAvatarURL(newAvatarURL)

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						gomega.Expect(*user.DisplayName).To(gomega.Equal(newDisplayName))
						gomega.Expect(*user.AvatarURL).To(gomega.Equal(newAvatarURL))
						return &updatedUser, nil
					})

				// Act
				result, err := useCase.Execute(ctx, userID, &newDisplayName, &newAvatarURL)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(*result.DisplayName).To(gomega.Equal(newDisplayName))
				gomega.Expect(*result.AvatarURL).To(gomega.Equal(newAvatarURL))
			})
		})

		ginkgo.Context("when updating only display name", func() {
			ginkgo.It("should update display name only", func() {
				// Arrange
				newDisplayName := "Updated Name"

				updatedUser := *testUser
				updatedUser.UpdateDisplayName(newDisplayName)

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						gomega.Expect(*user.DisplayName).To(gomega.Equal(newDisplayName))
						gomega.Expect(*user.AvatarURL).To(gomega.Equal("https://example.com/original.jpg")) // Should remain unchanged
						return &updatedUser, nil
					})

				// Act
				result, err := useCase.Execute(ctx, userID, &newDisplayName, nil)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(*result.DisplayName).To(gomega.Equal(newDisplayName))
			})
		})

		ginkgo.Context("when updating only avatar URL", func() {
			ginkgo.It("should update avatar URL only", func() {
				// Arrange
				newAvatarURL := "https://example.com/updated.jpg"

				updatedUser := *testUser
				updatedUser.UpdateAvatarURL(newAvatarURL)

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						gomega.Expect(*user.DisplayName).To(gomega.Equal("Original Name")) // Should remain unchanged
						gomega.Expect(*user.AvatarURL).To(gomega.Equal(newAvatarURL))
						return &updatedUser, nil
					})

				// Act
				result, err := useCase.Execute(ctx, userID, nil, &newAvatarURL)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(*result.AvatarURL).To(gomega.Equal(newAvatarURL))
			})
		})

		ginkgo.Context("when clearing display name with empty string", func() {
			ginkgo.It("should set display name to nil", func() {
				// Arrange
				emptyDisplayName := ""

				updatedUser := *testUser
				updatedUser.DisplayName = nil

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						gomega.Expect(user.DisplayName).To(gomega.BeNil())
						return &updatedUser, nil
					})

				// Act
				result, err := useCase.Execute(ctx, userID, &emptyDisplayName, nil)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.DisplayName).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when clearing avatar URL with empty string", func() {
			ginkgo.It("should set avatar URL to nil", func() {
				// Arrange
				emptyAvatarURL := ""

				updatedUser := *testUser
				updatedUser.AvatarURL = nil

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						gomega.Expect(user.AvatarURL).To(gomega.BeNil())
						return &updatedUser, nil
					})

				// Act
				result, err := useCase.Execute(ctx, userID, nil, &emptyAvatarURL)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.AvatarURL).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when user ID is nil", func() {
			ginkgo.It("should return error", func() {
				// Arrange
				newDisplayName := "Updated Name"

				// Act
				result, err := useCase.Execute(ctx, uuid.Nil, &newDisplayName, nil)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidUserID))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when user does not exist", func() {
			ginkgo.It("should return error", func() {
				// Arrange
				newDisplayName := "Updated Name"

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(nil, entity.ErrUserNotFound)

				// Act
				result, err := useCase.Execute(ctx, userID, &newDisplayName, nil)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrUserNotFound))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when display name is invalid", func() {
			ginkgo.It("should return error for too long display name", func() {
				// Arrange
				longDisplayName := string(make([]byte, 101)) // 101 characters, exceeds limit

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				// Act
				result, err := useCase.Execute(ctx, userID, &longDisplayName, nil)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidDisplayName))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when repository update fails", func() {
			ginkgo.It("should return repository error", func() {
				// Arrange
				newDisplayName := "Updated Name"

				mockRepo.EXPECT().
					GetByID(ctx, userID).
					Return(testUser, nil)

				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(nil, entity.ErrUserNotFound)

				// Act
				result, err := useCase.Execute(ctx, userID, &newDisplayName, nil)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrUserNotFound))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})
