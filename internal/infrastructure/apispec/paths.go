package apispec

// Base paths for gRPC services exposed by connect-go handlers
const (
	UserServiceBase   = "/user.v1.UserService/"
	HealthServiceBase = "/health.v1.HealthService/"
)

// ProtectedProcedures lists fully-qualified procedure paths that require authentication
var ProtectedProcedures = []string{
	UserServiceBase + "GetUserProfile",
	UserServiceBase + "UpdateUserProfile",
}
