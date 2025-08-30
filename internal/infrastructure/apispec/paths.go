package apispec

// Base paths for gRPC services exposed by connect-go handlers
const (
	UserServiceBase   = "/user.v1.UserService/"
	HealthServiceBase = "/health.v1.HealthService/"
)

// ProtectedProcedures returns a list of fully-qualified procedure paths that require authentication.
func ProtectedProcedures() []string {
	return []string{
		UserServiceBase + "GetUserProfile",
		UserServiceBase + "UpdateUserProfile",
	}
}
