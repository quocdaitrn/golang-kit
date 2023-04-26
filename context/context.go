package context

import "context"

// UID stores current user's identity.
type UID struct {
	Sub string `json:"user_id"`
	Tid string `json:"tid"`
}

func (u UID) IsZero() bool {
	return u.Sub == "" && u.Tid == ""
}

type uidKeyType struct{}

var uidKey = uidKeyType{}

func WithUID(ctx context.Context, uid UID) context.Context {
	return context.WithValue(ctx, uidKey, uid)
}

func UIDFromContext(ctx context.Context) UID {
	val := ctx.Value(uidKey)
	if uid, ok := val.(UID); ok {
		return uid
	}

	return UID{}
}
