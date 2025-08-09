package authctx

import "context"

type contextKey string

const (
	ClaimsKey contextKey = "claims"
)

// WithClaims adds claims to the context
func WithClaims(ctx context.Context, claims interface{}) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

// GetClaims retrieves claims from the context
func GetClaims[T any](ctx context.Context) *T {
	claims, _ := ctx.Value(ClaimsKey).(*T)
	return claims
}
