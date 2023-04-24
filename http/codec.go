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
