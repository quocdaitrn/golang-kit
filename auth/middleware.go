package auth

import (
	"context"
	"github.com/quocdaitrn/golang-kit/constant"
	"strings"

	"github.com/go-kit/kit/endpoint"

	kitcontext "github.com/quocdaitrn/golang-kit/context"
	kiterrors "github.com/quocdaitrn/golang-kit/errors"
)

type AuthenticateClient interface {
	IntrospectToken(ctx context.Context, accessToken string) (sub string, tid string, err error)
}

func Authenticate(ac AuthenticateClient) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			token, err := extractTokenFromHeaderString(constant.ContextAuthorization.Get(ctx))
			if err != nil {
				return nil, kiterrors.ErrUnauthorized.WithDetails(err)
			}

			sub, tid, err := ac.IntrospectToken(ctx, token)
			if err != nil {
				return nil, kiterrors.ErrUnauthorized.WithDetails(err)
			}

			uid := kitcontext.UID{
				Sub: sub,
				Tid: tid,
			}
			ctx = kitcontext.WithUID(ctx, uid)
			return next(ctx, request)
		}
	}
}

func extractTokenFromHeaderString(s string) (string, error) {
	parts := strings.Split(s, " ") //"Authorization" : "Bearer {token}"

	if parts[0] != "Bearer" || len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return "", kiterrors.ErrUnauthorized.WithDetails("missing access token")
	}

	return parts[1], nil
}

type errorUnauthorized struct {
	ErrorCode    int    `json:"_error"`
	ErrorMessage string `json:"_errorMessage"`
}
