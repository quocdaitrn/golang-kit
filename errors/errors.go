package errors

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err has stack inside, WithStack returns err only. If err is nil, WithStack
// returns nil.
//
// Example:
//
//	if err := DoSomething(); err != nil {
//		return errors.WithStack(err)
//	}
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	type formatter interface {
		Format(s fmt.State, verb rune)
	}
	if _, ok := err.(formatter); ok {
		return err
	}

	return errors.WithStack(err)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//	type causer interface {
//	       Cause() error
//	}
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.
func Cause(err error) error {
	return errors.Cause(err)
}

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(message string) error {
	return errors.New(message)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

// Error represents an error in kit.
type Error struct {
	// The internal code of this error.
	Code int `json:"code"`

	// The message of this error.
	Message string `json:"message,omitempty"`

	// The details of this error. Details usually is the error of the third party
	// lib, this field may include the stack trace, may be omitted in production
	// environment.
	Details interface{} `json:"details,omitempty"`

	// The message can be show for user.
	UserMessage string `json:"userMessage,omitempty" localize:"true"`
}

// Error returns error string, implements error interface.
func (e Error) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("Message: %s. Details: %#v.", e.Message, e.Details)
	}

	return e.Message
}

// NewError creates and returns a instance of Error.
func NewError(code int, msg string, details ...interface{}) *Error {
	err := &Error{Code: code, Message: msg}

	if len(details) > 0 {
		err.setDetails(details[0])
	}

	return err
}

// NewErrorWithUserMessage creates and returns a instance of Error with user
// message.
func NewErrorWithUserMessage(code int, msg string, userMsg string, details ...interface{}) *Error {
	err := &Error{Code: code, Message: msg, UserMessage: userMsg}

	if len(details) > 0 {
		err.setDetails(details[0])
	}

	return err
}

// setDetails sets and formats details of error.
func (e *Error) setDetails(details interface{}) {
	if err, ok := details.(error); ok {
		e.Details = err.Error()
	} else {
		e.Details = details
	}
}

// WithDetails returns a new Error with input details.
func (e Error) WithDetails(details interface{}) error {
	e.setDetails(details)
	return WithStack(&e)
}

// NotEqual reports whether this error not same with input error, on the
// internal code.
func (e Error) NotEqual(err error) bool {
	err = Cause(err)
	if ee, ok := err.(*Error); ok && ee.Code == e.Code {
		return false
	}

	return true
}

// Equal reports whether this error same with input error, on the internal
// code.
func (e Error) Equal(err error) bool {
	err = Cause(err)
	if ee, ok := err.(*Error); ok && ee.Code == e.Code {
		return true
	}

	return false
}

// Type returns error type.
func (e *Error) Type() string {
	return fmt.Sprintf("kit.errors.Error-%d-%s", e.Code, e.Message)
}

const (
	// List error in the Repository [0, 100).
	ErrCodeRepo = iota
	ErrCodeRepoInputValueMustBeArrayOrSlice
	ErrCodeRepoInputValueMustNotNil
	ErrCodeRepoInputValueMustNonEmpty
	ErrCodeRepoUnsupportedOperator
	ErrCodeRepoEntityNotFound
	ErrCodeRepoDuplicateKey
	ErrCodeRepoEntityIDUnspecified
	ErrCodeRepoIgnoreOp
	ErrCodeRepoCacheNotFound
	ErrCodeRepoCacheKeyEmpty
	ErrCodeRepoCacheInvalid

	// Common errors in services. [9000, 10000)
	ErrCodeInternalServerError = iota + 9000
	ErrCodeInvalidRequest
	ErrCodeRequestBindingFailed
	ErrCodeHTTPUnsupportedMediaType
	ErrCodeRequestScopesInvalid
	ErrCodeRequestSourceEmpty
	ErrCodeRequestSourceBlacklisted
	ErrCodeInsufficientPermission
	ErrCodeCloudEventsInvalidUnmarshalError
	ErrCodeClientError
	ErrCodeBadRequest
	ErrCodeUnauthorized
	ErrCodeAuthorizationHeaderMissing
	ErrCodeBearerTokenInvalid
	ErrCodeMultipleTokenProvided
	ErrCodeUnrecognizableToken
	ErrCodeTokenBlacklisted
	ErrCodeNoCredentialsMatch
	ErrCodeConsumerNotFound
	ErrCodeForbidden
	ErrCodeNotFound
	ErrCodeNotImplemented
	ErrCodeBadGateway
	ErrCodeServiceUnavailable
	ErrCodeRequestTimeout
	ErrCodeValidatorJSONSchemaNotFound
)

var businessErrors = map[int]bool{
	ErrCodeRepoEntityNotFound: true,
	ErrCodeRepoDuplicateKey:   true,
	ErrCodeRepoIgnoreOp:       true,

	ErrCodeInvalidRequest:             true,
	ErrCodeRequestBindingFailed:       true,
	ErrCodeHTTPUnsupportedMediaType:   true,
	ErrCodeRequestScopesInvalid:       true,
	ErrCodeRequestSourceEmpty:         true,
	ErrCodeRequestSourceBlacklisted:   true,
	ErrCodeInsufficientPermission:     true,
	ErrCodeClientError:                true,
	ErrCodeBadRequest:                 true,
	ErrCodeUnauthorized:               true,
	ErrCodeAuthorizationHeaderMissing: true,
	ErrCodeBearerTokenInvalid:         true,
	ErrCodeMultipleTokenProvided:      true,
	ErrCodeUnrecognizableToken:        true,
	ErrCodeTokenBlacklisted:           true,
	ErrCodeNoCredentialsMatch:         true,
	ErrCodeConsumerNotFound:           true,
	ErrCodeForbidden:                  true,
	ErrCodeNotFound:                   true,
	ErrCodeNotImplemented:             true,
	ErrCodeBadGateway:                 true,
	ErrCodeServiceUnavailable:         true,
	ErrCodeRequestTimeout:             true,
}

// IsBusinessError reports if input error is a business error.
func IsBusinessError(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}

	if ctx.Err() == context.Canceled {
		return true
	}

	err = Cause(err)
	switch e := err.(type) {
	case *Error:
		return businessErrors[e.Code]
	default:
		return false
	}
}

// AddBusinessError add a error code in to list of business errors.
func AddBusinessError(code int) {
	businessErrors[code] = true
}

// Error2KitError converts a input error to errors.Error.
func Error2KitError(err error) error {
	if err == nil {
		return nil
	}

	ke, ok := Cause(err).(*Error)
	if !ok {
		return ErrInternalServerError.WithDetails(err)
	}

	cpyKE := *ke
	return &cpyKE
}

// Errors in kit.
var (
	// ErrRepo represents for any undefined error which occur in the
	// repositories.
	ErrRepo = NewError(ErrCodeRepo, "something went wrong in the repository")

	// ErrRepoInputValueMustBeArrayOrSlice occurs when a input value is not an
	// array or slice while it must be. In query object of the repositories.
	ErrRepoInputValueMustBeArrayOrSlice = NewError(ErrCodeRepoInputValueMustBeArrayOrSlice, "the input value must be an array or a slice")

	// ErrRepoInputValueMustNotNil occurs when a input value is nil
	// while it must not. In query object of the repositories.
	ErrRepoInputValueMustNotNil = NewError(ErrCodeRepoInputValueMustNotNil, "the input must be not nil")

	// ErrRepoInputValueMustNonEmpty occurs when a input value is an empty
	// slice or array while it must not. In query object of the
	// repositories.
	ErrRepoInputValueMustNonEmpty = NewError(ErrCodeRepoInputValueMustNonEmpty, "the input must be a non empty array or slice")

	// ErrRepoUnsupportedOperator occurs when an operator of a query expression
	// is unsupported by a specific repo.
	ErrRepoUnsupportedOperator = NewError(ErrCodeRepoUnsupportedOperator, "the operator is not supported")

	// ErrRepoEntityNotFound occurs when finding an entity that not exists in
	// repository.
	ErrRepoEntityNotFound = NewError(ErrCodeRepoEntityNotFound, "entity not found")

	// ErrRepoDuplicateKey occurs when trying to insert an entity to repository
	// with an existed id.
	ErrRepoDuplicateKey = NewError(ErrCodeRepoDuplicateKey, "duplicate key error")

	// ErrRepoEntityIDUnspecified is error for inserting an entity to repository
	// with an id empty.
	ErrRepoEntityIDUnspecified = NewError(ErrCodeRepoEntityIDUnspecified, "the id field of entity must be specified")

	// ErrRepoIgnoreOp is error when a op is ignored because a previous op is error.
	ErrRepoIgnoreOp = NewError(ErrCodeRepoIgnoreOp, "this op is ignored because a previous op is error")

	// ErrRepoCacheNotFound is error when finding a item that is not existed in
	// cache.
	ErrRepoCacheNotFound = NewError(ErrCodeRepoCacheNotFound, "not found in cache")

	// ErrRepoCacheKeyEmpty is error when setting or getting from cache by an
	// empty key.
	ErrRepoCacheKeyEmpty = NewError(ErrCodeRepoCacheKeyEmpty, "the key must be specific")

	// ErrRepoCacheInvalid is error when getting an invalid item from cache.
	ErrRepoCacheInvalid = NewError(ErrCodeRepoCacheInvalid, "the item is invalid")

	// ErrInternalServerError is error for common error in server.
	ErrInternalServerError = NewError(ErrCodeInternalServerError, "oops! something went wrong")

	// ErrInvalidRequest is error when request form is invalid in endpoint.
	ErrInvalidRequest = NewError(ErrCodeInvalidRequest, "invalid request")

	// ErrRequestScopesInvalid is error when a request make without a valid
	// scopes.
	ErrRequestScopesInvalid = NewError(ErrCodeRequestScopesInvalid, "invalid request scopes")

	// ErrInsufficientPermission is an error which occurres when a request does
	// not have permission to access a action.
	ErrInsufficientPermission = NewError(ErrCodeInsufficientPermission, "insufficient permission")

	// ErrRequestBindingFailed is error when bind http request fail.
	ErrRequestBindingFailed = NewError(ErrCodeRequestBindingFailed, "failed to bind request")

	// ErrHTTPUnsupportedMediaType is error when bind http request fail.
	ErrHTTPUnsupportedMediaType = NewError(ErrCodeHTTPUnsupportedMediaType, "unsupported media type")

	// ErrCloudEventsInvalidUnmarshalError is error when bind http request
	// fail.
	ErrCloudEventsInvalidUnmarshalError = NewError(ErrCodeCloudEventsInvalidUnmarshalError, "invalid CloudEvents unmarshal")

	// ErrRequestSourceEmpty is error when a request made with a empty
	// source.
	ErrRequestSourceEmpty = NewError(ErrCodeRequestSourceEmpty, "request source is empty")

	// ErrRequestSourceBlacklisted is error when a request made without a empty
	// source.
	ErrRequestSourceBlacklisted = NewError(ErrCodeRequestSourceBlacklisted, "request source is blacklisted")

	// ErrClientError represents a common client error. It usual 4xx errors in
	// HTTP.
	ErrClientError = NewError(ErrCodeClientError, "client error")

	// ErrBadRequest represents a bad request.
	ErrBadRequest = NewError(ErrCodeBadRequest, "bad request")

	// ErrUnauthorized is a common authorization error.
	ErrUnauthorized = NewError(ErrCodeUnauthorized, "unauthorized")

	// ErrAuthorizationHeaderMissing is an error which occurres when
	// missing 'Authorization' header in the request.
	ErrAuthorizationHeaderMissing = NewError(ErrCodeAuthorizationHeaderMissing, "authorization header missing")

	// ErrBearerTokenInvalid is an error which occurres when the token
	// isn't a Bearer token.
	ErrBearerTokenInvalid = NewError(ErrCodeBearerTokenInvalid, "invalid bearer token")

	// ErrMultipleTokenProvied is an error which occurres when there are
	// more than one token in the request.
	ErrMultipleTokenProvied = NewError(ErrCodeMultipleTokenProvided, "multiple tokens provided")

	// ErrUnrecognizableToken is an error which occurres when the token
	// unable to recognizable.
	ErrUnrecognizableToken = NewError(ErrCodeUnrecognizableToken, "unrecognizable token")

	// ErrTokenBlacklisted is an error which occurres when the token is
	// blacklisted.
	ErrTokenBlacklisted = NewError(ErrCodeTokenBlacklisted, "the token has been blacklisted")

	// ErrForbidden is a common forbidden error.
	ErrForbidden = NewError(ErrCodeForbidden, "forbidden")

	// ErrNotFound is a common error not found.
	ErrNotFound = NewError(ErrCodeNotFound, "not found")

	// ErrNotImplemented is a common error not implemented.
	ErrNotImplemented = NewError(ErrCodeNotImplemented, "not implemented")

	// ErrBadGateway is a common error bad gateway.
	ErrBadGateway = NewError(ErrCodeBadGateway, "bad gateway")

	// ErrServiceUnavailable is a common error service unavailable.
	ErrServiceUnavailable = NewError(ErrCodeServiceUnavailable, "service unavailable")

	// ErrRequestTimeout is a common error gateway timeout.
	ErrRequestTimeout = NewError(ErrCodeRequestTimeout, "request timeout")

	// ErrValidatorJSONSchemaNotFound is a common error valodator JSON schema not found.
	ErrValidatorJSONSchemaNotFound = NewError(ErrCodeValidatorJSONSchemaNotFound, "validator JSON schema not found")
)
