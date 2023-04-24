package errors

import (
	"fmt"
	"net/http"

	kiterrors "github.com/quocdaitrn/golang-kit/errors"
)

// HTTPError defines a HTTP error that can be returned
// in a Response from the spec
type HTTPError struct {
	HTTPStatus int         `json:"-"`
	Code       int         `json:"_error"`
	Message    string      `json:"_errorMessage"`
	Details    interface{} `json:"_errorDetails,omitempty"`

	UserMessage string `json:"_userMessage,omitempty"`
}

// NewHTTPError creates, initializes and returns a new HTTPError instance.
func NewHTTPError(httpStatus int, code int, msg string, details ...interface{}) *HTTPError {
	fe := &HTTPError{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    msg,
		Details:    nil,
	}
	if len(details) > 0 {
		fe.Details = details[0]
	}

	return fe
}

// HTTPErr returns http error itself.
func (e *HTTPError) HTTPErr() *HTTPError {
	return e
}

// Error returns error string, implements error interface.
func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTPStatus: %d, Code: %d, Message: %s, Details: %#v.", e.HTTPStatus, e.Code, e.Message, e.Details)
}

// Type returns error type.
func (e HTTPError) Type() string {
	return fmt.Sprintf("http.Error%d", e.Code)
}

// Error2HTTPError converts a input error to errors.HTTPError.
func Error2HTTPError(err error) error {
	if err == nil {
		return nil
	}

	ke, ok := kiterrors.Cause(err).(*kiterrors.Error)
	if !ok {
		return HTTPErrInternalServerError
	}

	he, ok := kitError2HTTPErrorMapping[ke.Code]
	if !ok {
		return HTTPErrInternalServerError
	}
	he.Details = ke.Details
	he.UserMessage = ke.UserMessage

	return &he
}

// HTTPError2KitError converts an errors.HTTPError to kit Error.
func HTTPError2KitError(he *HTTPError) error {
	if he == nil {
		return nil
	}

	ke, ok := httpError2KitErrorMapping[he.Code]
	if !ok {
		switch {
		case he.HTTPStatus >= 400 && he.HTTPStatus < 500:
			return kiterrors.ErrClientError
		default:
			return kiterrors.ErrInternalServerError
		}
	}
	ke.Details = he.Details
	ke.UserMessage = he.UserMessage

	return &ke
}

// kitError2HTTPErrorMapping maps an internal error code to http error.
var kitError2HTTPErrorMapping = map[int]HTTPError{
	kiterrors.ErrCodeRepoEntityNotFound:         *HTTPErrResourceNotFound,
	kiterrors.ErrCodeRepoDuplicateKey:           *HTTPErrDuplicateKey,
	kiterrors.ErrCodeRepoIgnoreOp:               *HTTPErrRepoIgnoreOp,
	kiterrors.ErrCodeInvalidRequest:             *HTTPErrInvalidRequest,
	kiterrors.ErrCodeInsufficientPermission:     *HTTPErrInsufficientPermission,
	kiterrors.ErrCodeRequestScopesInvalid:       *HTTPErrInsufficientPermission,
	kiterrors.ErrCodeRequestSourceEmpty:         *HTTPErrInsufficientPermission,
	kiterrors.ErrCodeRequestBindingFailed:       *HTTPErrRequestBindingFailed,
	kiterrors.ErrCodeBadRequest:                 *HTTPErrBadRequest,
	kiterrors.ErrCodeUnauthorized:               *HTTPErrUnauthorized,
	kiterrors.ErrCodeAuthorizationHeaderMissing: *HTTPErrAuthorizationHeaderMissing,
	kiterrors.ErrCodeBearerTokenInvalid:         *HTTPErrBearerTokenInvalid,
	kiterrors.ErrCodeMultipleTokenProvided:      *HTTPErrMultipleTokenProvied,
	kiterrors.ErrCodeUnrecognizableToken:        *HTTPErrUnrecognizableToken,
	kiterrors.ErrCodeTokenBlacklisted:           *HTTPErrTokenBlacklisted,
	kiterrors.ErrCodeForbidden:                  *HTTPErrForbidden,
	kiterrors.ErrCodeNotFound:                   *HTTPErrNotFound,
	kiterrors.ErrCodeNotImplemented:             *HTTPErrNotImplemented,
	kiterrors.ErrCodeBadGateway:                 *HTTPErrBadGateway,
	kiterrors.ErrCodeServiceUnavailable:         *HTTPErrServiceUnavailable,
	kiterrors.ErrCodeRequestTimeout:             *HTTPErrGatewayTimeout,
}

var httpError2KitErrorMapping = map[int]kiterrors.Error{
	HTTPErrResourceNotFound.Code:           *kiterrors.ErrRepoEntityNotFound,
	HTTPErrDuplicateKey.Code:               *kiterrors.ErrRepoDuplicateKey,
	HTTPErrRepoIgnoreOp.Code:               *kiterrors.ErrRepoIgnoreOp,
	HTTPErrInvalidRequest.Code:             *kiterrors.ErrInvalidRequest,
	HTTPErrInsufficientPermission.Code:     *kiterrors.ErrInsufficientPermission,
	HTTPErrRequestBindingFailed.Code:       *kiterrors.ErrRequestBindingFailed,
	HTTPErrBadRequest.Code:                 *kiterrors.ErrBadRequest,
	HTTPErrUnauthorized.Code:               *kiterrors.ErrUnauthorized,
	HTTPErrAuthorizationHeaderMissing.Code: *kiterrors.ErrAuthorizationHeaderMissing,
	HTTPErrBearerTokenInvalid.Code:         *kiterrors.ErrBearerTokenInvalid,
	HTTPErrMultipleTokenProvied.Code:       *kiterrors.ErrMultipleTokenProvied,
	HTTPErrUnrecognizableToken.Code:        *kiterrors.ErrUnrecognizableToken,
	HTTPErrTokenBlacklisted.Code:           *kiterrors.ErrTokenBlacklisted,
	HTTPErrForbidden.Code:                  *kiterrors.ErrForbidden,
	HTTPErrNotFound.Code:                   *kiterrors.ErrNotFound,
	HTTPErrNotImplemented.Code:             *kiterrors.ErrNotImplemented,
	HTTPErrBadGateway.Code:                 *kiterrors.ErrBadGateway,
	HTTPErrServiceUnavailable.Code:         *kiterrors.ErrServiceUnavailable,
	HTTPErrGatewayTimeout.Code:             *kiterrors.ErrRequestTimeout,
}

// AddError2HTTPErrorMapping add a mapping internal code to
// http.errors.HTTPError.
func AddError2HTTPErrorMapping(code int, he *HTTPError) {
	kitError2HTTPErrorMapping[code] = *he
}

// AddErrorHTTPErrorMapping add a mapping internal Error and
// http.errors.HTTPError.
func AddErrorHTTPErrorMapping(ke *kiterrors.Error, he *HTTPError) {
	kitError2HTTPErrorMapping[ke.Code] = *he
	httpError2KitErrorMapping[he.Code] = *ke
}

// Commmon HTTP Errors used in services.
var (
	// HTTPErrBadRequest is a common bad request error.
	HTTPErrBadRequest = NewHTTPError(http.StatusBadRequest, 400000, "Bad request")

	// HTTPErrValidationFormFailed is an error when failed to validate request.
	HTTPErrInvalidRequest = NewHTTPError(http.StatusBadRequest, 400002, "Input form is invalid")

	// HTTPErrRequestBindingFailed is an error when failed to bind http
	// request.
	HTTPErrRequestBindingFailed = NewHTTPError(http.StatusBadRequest, 400003, "Cannot binding request")

	// HTTPErrUnauthorized is a common authorization error.
	HTTPErrUnauthorized = NewHTTPError(http.StatusUnauthorized, 401000, "Unauthorized")

	// HTTPErrAuthorizationHeaderMissing is an error which occurres when
	// missing 'Authorization' header in the request.
	HTTPErrAuthorizationHeaderMissing = NewHTTPError(http.StatusUnauthorized, 401001, "Missing authorization header")

	// HTTPErrBearerTokenInvalid is an error which occurres when the token
	// isn't a Bearer token.
	HTTPErrBearerTokenInvalid = NewHTTPError(http.StatusUnauthorized, 401002, "Bad token, invalid bearer token")

	// HTTPErrMultipleTokenProvied is an error which occurres when there are
	// more than one token in the request.
	HTTPErrMultipleTokenProvied = NewHTTPError(http.StatusUnauthorized, 401003, "Multiple tokens provided")

	// HTTPErrUnrecognizableToken is an error which occurres when the token
	// unable to recognizable.
	HTTPErrUnrecognizableToken = NewHTTPError(http.StatusUnauthorized, 401004, "Unrecognizable token")

	// HTTPErrTokenBlacklisted is an error which occurres when the token is
	// blacklisted.
	HTTPErrTokenBlacklisted = NewHTTPError(http.StatusUnauthorized, 401005, "The token has been blacklisted")

	// HTTPErrInsufficientPermission is an error which occurres when a request
	// does not have permission to access a API.
	HTTPErrInsufficientPermission = NewHTTPError(http.StatusForbidden, 403050, "Insufficient Permission")

	// HTTPErrForbidden is a common error forbidden.
	HTTPErrForbidden = NewHTTPError(http.StatusForbidden, 403000, "Forbidden")

	// HTTPErrNotFound is a common error not found.
	HTTPErrNotFound = NewHTTPError(http.StatusNotFound, 404000, "Not found")

	// HTTPErrResourceNotFound is error for querying the database for a document
	// which does not exist.
	HTTPErrResourceNotFound = NewHTTPError(http.StatusNotFound, 404001, "Resource not found")

	// HTTPErrDuplicateKey is error for inserting a document to database with
	// a duplicate id or values which are marked as unique index.
	HTTPErrDuplicateKey = NewHTTPError(http.StatusConflict, 409001, "Duplicate key error")

	// HTTPErrInternalServerError is common internal error in server.
	HTTPErrInternalServerError = NewHTTPError(http.StatusInternalServerError, 500000, "Oops, something went wrong")

	// HTTPErrRetrieveTokenFailed is an authorization when server failed to get
	// Bearer token from the value of header 'Authorization'.
	HTTPErrRetrieveTokenFailed = NewHTTPError(http.StatusInternalServerError, 500002, "Cannot retrieve token")

	// HTTPErrRepoIgnoreOp is error when a op is ignored because a previous op is error.
	HTTPErrRepoIgnoreOp = NewHTTPError(http.StatusInternalServerError, 500201, "Oops, an error occurred with Repository")

	// HTTPErrNotImplemented is common error not implemented.
	HTTPErrNotImplemented = NewHTTPError(http.StatusNotImplemented, 501000, "Not implemented")

	// HTTPErrBadGateway is common error bad gateway.
	HTTPErrBadGateway = NewHTTPError(http.StatusBadGateway, 502000, "Bad gateway")

	// HTTPErrServiceUnavailable is common error service unavailable.
	HTTPErrServiceUnavailable = NewHTTPError(http.StatusServiceUnavailable, 503000, "Service unavailable")

	// HTTPErrGatewayTimeout is common error service unavailable.
	HTTPErrGatewayTimeout = NewHTTPError(http.StatusGatewayTimeout, 504000, "Gateway timeout")
)
