package constant

import "context"

const (
	// ContextIsDebug is used for storing isDebug flag for service.
	ContextIsDebug contextBool = iota
)

type contextBool int

// WithValue returns a copy of parent in which the value associated with key is
// val.
func (k contextBool) WithValue(ctx context.Context, v bool) context.Context {
	return context.WithValue(ctx, k, v)
}

// Get returns the value from context.
func (k contextBool) Get(ctx context.Context) bool {
	if ctx.Value(k) == nil {
		return false
	}

	return true
}

const (
	// ContextAuthorization is used for storing authorization token of user.
	ContextAuthorization contextString = iota
)

type contextString int

// WithValue returns a copy of parent in which the value associated with key is
// val.
func (k contextString) WithValue(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, k, v)
}

// Get returns the value from context.
func (k contextString) Get(ctx context.Context) string {
	v := ctx.Value(k)
	if v == nil {
		return ""
	}

	return v.(string)
}
