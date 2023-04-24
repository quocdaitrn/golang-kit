package http

import (
	"context"
	"net/http"

	kitconstant "github.com/quocdaitrn/golang-kit/constant"
	kiterrors "github.com/quocdaitrn/golang-kit/errors"
	httperrors "github.com/quocdaitrn/golang-kit/http/errors"
)

func DefaultErrorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	isDebug := kitconstant.ContextIsDebug.Get(ctx)
	e := httperrors.Error2HTTPError(err)
	if e == nil {
		return
	}

	he := e.(*httperrors.HTTPError)
	if !kiterrors.IsBusinessError(ctx, err) && !isDebug {
		he.Details = nil
		he.Message = "Oops, something went wrong"
	}

	contentType := "application/json; charset=utf-8"
	w.Header().Set("Content-Type", contentType)

	w.WriteHeader(he.HTTPStatus)

	if err := json.NewEncoder(w).Encode(he); err != nil {
		// TODO: should handle error here.
	}
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// EncodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		DefaultErrorEncoder(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
