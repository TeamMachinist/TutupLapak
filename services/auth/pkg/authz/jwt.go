package authz

import (
	"context"
)

// Claims - shared structure for responses (no JWT logic here)
type Claims struct {
	UserID string `json:"user_id"`
}

// Context helpers
type contextKey string

const UserIDKey contextKey = "user_id"

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
