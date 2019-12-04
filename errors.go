package sreq

import (
	"errors"
	"fmt"
)

var (
	// ErrUnexpectedTransport can be used if assert a RoundTripper as a non-nil *http.Transport instance failed.
	ErrUnexpectedTransport = errors.New("current transport isn't a non-nil *http.Transport instance")

	// ErrNilContext can be used when the context is nil.
	ErrNilContext = errors.New("nil Context")

	// ErrNilCookieJar can be used when the cookie jar is nil.
	ErrNilCookieJar = errors.New("sreq: nil cookie jar")

	// ErrJarCookiesNotPresent can be used when cookies for a given URL not present in cookie jar.
	ErrJarCookiesNotPresent = errors.New("sreq: cookies for the given URL not present")

	// ErrJarNamedCookieNotPresent can be used when named cookie for a given URL not present in cookie jar.
	ErrJarNamedCookieNotPresent = errors.New("sreq: named cookie for the given URL not present")

	// ErrResponseCookiesNotPresent can be used when cookies of the HTTP response not present.
	ErrResponseCookiesNotPresent = errors.New("sreq: cookies not present")

	// ErrResponseNamedCookieNotPresent can be used when named cookie of the HTTP response not present.
	ErrResponseNamedCookieNotPresent = errors.New("sreq: named cookie not present")

	// ErrRetryMaxDurationExceeded can be used when the retry max duration exceeded.
	ErrRetryMaxDurationExceeded = errors.New("sreq: retry max duration exceeded")
)

type (
	// ClientError records a client error, can be used when sreq builds Client failed.
	ClientError struct {
		Cause string
		Err   error
	}

	// RequestError records a request error, can be used when sreq builds Request failed.
	RequestError struct {
		Cause string
		Err   error
	}
)

// Error implements error interface.
func (c *ClientError) Error() string {
	return fmt.Sprintf("sreq>>Client [%s]: %s", c.Cause, c.Err.Error())
}

// Unwrap unpacks and returns the wrapped err of c.
func (c *ClientError) Unwrap() error {
	return c.Err
}

// Error implements error interface.
func (req *RequestError) Error() string {
	return fmt.Sprintf("sreq>>Request [%s]: %s", req.Cause, req.Err.Error())
}

// Unwrap unpacks and returns the wrapped err of req.
func (req *RequestError) Unwrap() error {
	return req.Err
}
