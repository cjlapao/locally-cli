package auth

import (
	"context"
)

type contextKey string

const (
	ClaimsKey contextKey = "claims"
)

// WithClaims adds claims to the context
func WithClaims(ctx context.Context, claims *AuthClaims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

// GetClaims retrieves claims from the context
func GetClaims(ctx context.Context) *AuthClaims {
	claims, _ := ctx.Value(ClaimsKey).(*AuthClaims)
	return claims
}
