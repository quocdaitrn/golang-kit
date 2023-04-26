package http

import (
	"context"
	"net/http"
)

const RequestHeaderAuthorization = "Authorization"

// PopulateRequestAuthorizationToken is a RequestFunc that populates scopes values
// from "Authorization" header to the context.
func PopulateRequestAuthorizationToken(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, RequestHeaderAuthorization, r.Header.Get(HeaderAuthorization))
}
