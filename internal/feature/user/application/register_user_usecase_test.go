package application_test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/feature/user/application"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/mocks"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("RegisterUserUseCase", func() {
	var (
		ctrl         *gomock.Controller
		mockRepo     *mocks.MockUserRepository
		useCase      application.RegisterUserUseCase
		ctx          context.Context
		firebaseUID  string
		displayName  string
		authProvider string
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockRepo = mocks.NewMockUserRepository(ctrl)
		useCase = application.NewRegisterUserUseCase(mockRepo)
		ctx = context.Background()
		firebaseUID = "firebase_test_uid"
		displayName = "Test User"
		authProvider = "firebase"
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("Execute", func() {
		ginkgo.Context("when all parameters are valid", func() {
			ginkgo.It("should create user successfully", func() {
				// Arrange
				expectedUser := entity.NewUser(firebaseUID, authProvider)
				expectedUser.UpdateDisplayName(displayName)

				mockRepo.EXPECT().
					Create(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						gomega.Expect(user.FirebaseUID).To(gomega.Equal(firebaseUID))
						gomega.Expect(*user.DisplayName).To(gomega.Equal(displayName))
						gomega.Expect(user.AuthProvider).To(gomega.Equal(authProvider))
						return expectedUser, nil
					})

				// Act
				result, err := useCase.Execute(ctx, firebaseUID, displayName, authProvider)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.FirebaseUID).To(gomega.Equal(firebaseUID))
				gomega.Expect(*result.DisplayName).To(gomega.Equal(displayName))
				gomega.Expect(result.AuthProvider).To(gomega.Equal(authProvider))
			})
		})

		ginkgo.Context("when display name is empty", func() {
			ginkgo.It("should create user without display name", func() {
				// Arrange
				expectedUser := entity.NewUser(firebaseUID, authProvider)

				mockRepo.EXPECT().
					Create(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, user *entity.User) (*entity.User, error) {
						gomega.Expect(user.FirebaseUID).To(gomega.Equal(firebaseUID))
						gomega.Expect(user.DisplayName).To(gomega.BeNil())
						gomega.Expect(user.AuthProvider).To(gomega.Equal(authProvider))
						return expectedUser, nil
					})

				// Act
				result, err := useCase.Execute(ctx, firebaseUID, "", authProvider)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.FirebaseUID).To(gomega.Equal(firebaseUID))
				gomega.Expect(result.DisplayName).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when firebase UID is empty", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := useCase.Execute(ctx, "", displayName, authProvider)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidFirebaseUID))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when firebase UID is whitespace only", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := useCase.Execute(ctx, "   ", displayName, authProvider)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidFirebaseUID))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when auth provider is empty", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := useCase.Execute(ctx, firebaseUID, displayName, "")

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidAuthProvider))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when auth provider is whitespace only", func() {
			ginkgo.It("should return error", func() {
				// Act
				result, err := useCase.Execute(ctx, firebaseUID, displayName, "   ")

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidAuthProvider))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when display name is invalid", func() {
			ginkgo.It("should return error for too long display name", func() {
				// Arrange
				longDisplayName := string(make([]byte, 101)) // 101 characters, exceeds limit

				// Act
				result, err := useCase.Execute(ctx, firebaseUID, longDisplayName, authProvider)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrInvalidDisplayName))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when repository returns error", func() {
			ginkgo.It("should return repository error", func() {
				// Arrange
				mockRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil, entity.ErrUserAlreadyExists)

				// Act
				result, err := useCase.Execute(ctx, firebaseUID, displayName, authProvider)

				// Assert
				gomega.Expect(err).To(gomega.Equal(entity.ErrUserAlreadyExists))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})
