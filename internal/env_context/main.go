package envcontext

import (
	"context"
)

type EnvironmentContext struct {
	ctx context.Context
}

func WithContext(ctx context.Context) *EnvironmentContext {
	return &EnvironmentContext{
		ctx: ctx,
	}
}

func (ec *EnvironmentContext) Context() context.Context {
	return ec.ctx
}

func (ec *EnvironmentContext) SetValue(key, value interface{}) *EnvironmentContext {
	ec.ctx = context.WithValue(ec.ctx, key, value)
	return ec
}
