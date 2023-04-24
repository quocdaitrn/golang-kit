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
