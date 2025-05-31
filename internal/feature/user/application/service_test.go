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

var _ = ginkgo.Describe("UserService", func() {
	var (
		ctrl         *gomock.Controller
		mockRepo     *mocks.MockUserRepository
		userService  application.UserService
		ctx          context.Context
		testUser     *entity.User
		testUserID   uuid.UUID
		firebaseUID  string
		displayName  string
		authProvider string
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockRepo = mocks.NewMockUserRepository(ctrl)
		userService = application.NewService(mockRepo)
		ctx = context.Background()

		testUserID = uuid.New()
		firebaseUID = "firebase_test_uid"
		displayName = "Test User"
		authProvider = "firebase"
		testUser = entity.NewUser(firebaseUID, authProvider)
		testUser.ID = testUserID
		testUser.UpdateDisplayName(displayName)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("RegisterUser", func() {
		ginkgo.Context("when registration is successful", func() {
			ginkgo.It("should register user successfully", func() {
				// Arrange
				mockRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(testUser, nil)

				// Act
				result, err := userService.RegisterUser(ctx, firebaseUID, displayName, authProvider)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.FirebaseUID).To(gomega.Equal(firebaseUID))
				gomega.Expect(*result.DisplayName).To(gomega.Equal(displayName))
				gomega.Expect(result.AuthProvider).To(gomega.Equal(authProvider))
			})
		})

		ginkgo.Context("when user already exists", func() {
			ginkgo.It("should return error", func() {
				// Arrange
				mockRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, entity.ErrUserAlreadyExists)

				// Act
				result, err := userService.RegisterUser(ctx, firebaseUID, displayName, authProvider)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrUserAlreadyExists))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when firebase UID is invalid", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := userService.RegisterUser(ctx, "", displayName, authProvider)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidFirebaseUID))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when auth provider is invalid", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := userService.RegisterUser(ctx, firebaseUID, displayName, "")

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidAuthProvider))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("GetUserProfile", func() {
		ginkgo.Context("when user exists", func() {
			ginkgo.It("should return user profile", func() {
				// Arrange
				mockRepo.EXPECT().
					GetByID(ctx, testUserID).
					Return(testUser, nil)

				// Act
				result, err := userService.GetUserProfile(ctx, testUserID)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.ID).To(gomega.Equal(testUserID))
				gomega.Expect(result.FirebaseUID).To(gomega.Equal(firebaseUID))
			})
		})

		ginkgo.Context("when user does not exist", func() {
			ginkgo.It("should return error", func() {
				// Arrange
				mockRepo.EXPECT().
					GetByID(ctx, testUserID).
					Return(nil, entity.ErrUserNotFound)

				// Act
				result, err := userService.GetUserProfile(ctx, testUserID)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrUserNotFound))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when user ID is invalid", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := userService.GetUserProfile(ctx, uuid.Nil)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidUserID))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("UpdateUserProfile", func() {
		var (
			newDisplayName string
			newAvatarURL   string
		)

		ginkgo.BeforeEach(func() {
			newDisplayName = "Updated User"
			newAvatarURL = "https://example.com/avatar.jpg"
		})

		ginkgo.Context("when update is successful", func() {
			ginkgo.It("should update user profile", func() {
				// Arrange
				updatedUser := *testUser
				updatedUser.UpdateDisplayName(newDisplayName)
				updatedUser.UpdateAvatarURL(newAvatarURL)

				mockRepo.EXPECT().
					GetByID(ctx, testUserID).
					Return(testUser, nil)

				mockRepo.EXPECT().
					Update(ctx, gomock.Any()).
					Return(&updatedUser, nil)

				// Act
				result, err := userService.UpdateUserProfile(ctx, testUserID, &newDisplayName, &newAvatarURL)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(*result.DisplayName).To(gomega.Equal(newDisplayName))
				gomega.Expect(*result.AvatarURL).To(gomega.Equal(newAvatarURL))
			})
		})

		ginkgo.Context("when user does not exist", func() {
			ginkgo.It("should return error", func() {
				// Arrange
				mockRepo.EXPECT().
					GetByID(ctx, testUserID).
					Return(nil, entity.ErrUserNotFound)

				// Act
				result, err := userService.UpdateUserProfile(ctx, testUserID, &newDisplayName, &newAvatarURL)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrUserNotFound))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when user ID is invalid", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := userService.UpdateUserProfile(ctx, uuid.Nil, &newDisplayName, &newAvatarURL)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidUserID))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})
