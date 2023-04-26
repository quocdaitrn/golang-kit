package http

import (
	"context"
	"net/http"

	"github.com/quocdaitrn/golang-kit/constant"
)

// PopulateRequestAuthorizationToken is a RequestFunc that populates scopes values
// from "Authorization" header to the context.
func PopulateRequestAuthorizationToken(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, constant.ContextAuthorization, r.Header.Get(HeaderAuthorization))
}
