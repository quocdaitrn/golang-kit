package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	kitcontext "github.com/quocdaitrn/golang-kit/context"
	kiterrors "github.com/quocdaitrn/golang-kit/errors"
)

type AuthenticateClient interface {
	IntrospectToken(ctx context.Context, accessToken string) (sub string, tid string, err error)
}

func RequireAuth(ac AuthenticateClient) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractTokenFromHeaderString(r.Header.Get("Authorization"))
			if err != nil {
				unauthorizedErr, _ := json.Marshal(&errorUnauthorized{
					ErrorCode:    401001,
					ErrorMessage: err.Error(),
				})
				http.Error(w, string(unauthorizedErr), http.StatusUnauthorized)
				return
			}

			sub, tid, err := ac.IntrospectToken(r.Context(), token)
			if err != nil {
				unauthorizedErr, _ := json.Marshal(&errorUnauthorized{
					ErrorCode:    401000,
					ErrorMessage: "unauthorized",
				})
				http.Error(w, string(unauthorizedErr), http.StatusUnauthorized)
				return
			}

			uid := kitcontext.UID{
				Sub: sub,
				Tid: tid,
			}
			ctx := kitcontext.WithUID(r.Context(), uid)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
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
