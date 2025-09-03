package application

import "context"

// context key type to avoid collisions
type ensureUserCtxKey string

const initialDisplayNameKey ensureUserCtxKey = "initial_display_name"

// WithInitialDisplayName injects an initial display name into context for EnsureUser use case.
func WithInitialDisplayName(ctx context.Context, displayName string) context.Context {
    if displayName == "" {
        return ctx
    }
    return context.WithValue(ctx, initialDisplayNameKey, displayName)
}

// initialDisplayNameFromContext retrieves an optional initial display name from context.
func initialDisplayNameFromContext(ctx context.Context) (string, bool) {
    val := ctx.Value(initialDisplayNameKey)
    if s, ok := val.(string); ok && s != "" {
        return s, true
    }
    return "", false
}

