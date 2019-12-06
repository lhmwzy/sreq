package sreq

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	stdurl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MethodGet represents the GET method for HTTP.
	MethodGet = "GET"

	// MethodHead represents the HEAD method for HTTP.
	MethodHead = "HEAD"

	// MethodPost represents the POST method for HTTP.
	MethodPost = "POST"

	// MethodPut represents the PUT method for HTTP.
	MethodPut = "PUT"

	// MethodPatch represents the PATCH method for HTTP.
	MethodPatch = "PATCH"

	// MethodDelete represents the DELETE method for HTTP.
	MethodDelete = "DELETE"

	// MethodConnect represents the CONNECT method for HTTP.
	MethodConnect = "CONNECT"

	// MethodOptions represents the OPTIONS method for HTTP.
	MethodOptions = "OPTIONS"

	// MethodTrace represents the TRACE method for HTTP.
	MethodTrace = "TRACE"
)

type (
	// Request wraps the raw HTTP request.
	Request struct {
		RawRequest *http.Request
		Host       string
		Headers    Headers
		Cookies    []*http.Cookie
		Timeout    time.Duration
		Err        error

		auth        *auth
		bearerToken string
		ctx         context.Context
		retry       *retry
	}

	// RequestOption specifies a request options, like params, form, etc.
	RequestOption func(*Request) *Request

	// RequestInterceptor specifies a request interceptor.
	RequestInterceptor func(*Request) error
)

func (req *Request) raiseError(cause string, err error) {
	req.Err = &RequestError{
		Cause: cause,
		Err:   err,
	}
}

// NewRequest returns a new Request given a method, URL.
func NewRequest(method string, url string) *Request {
	req := &Request{
		Headers: make(Headers),
	}

	rawRequest, err := http.NewRequest(method, url, nil)
	if err != nil {
		req.raiseError("NewRequest", err)
		return req
	}

	rawRequest.Header.Set("User-Agent", "sreq "+Version)
	req.RawRequest = rawRequest
	return req
}

// Raw returns the raw HTTP request.
func (req *Request) Raw() (*http.Request, error) {
	return req.RawRequest, req.Err
}

// SetBody sets body for the HTTP request.
func (req *Request) SetBody(body io.Reader) *Request {
	if req.Err != nil {
		return req
	}

	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = ioutil.NopCloser(body)
	}
	req.RawRequest.Body = rc

	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			req.RawRequest.ContentLength = int64(v.Len())
			buf := v.Bytes()
			req.RawRequest.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return ioutil.NopCloser(r), nil
			}
		case *bytes.Reader:
			req.RawRequest.ContentLength = int64(v.Len())
			snapshot := *v
			req.RawRequest.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		case *strings.Reader:
			req.RawRequest.ContentLength = int64(v.Len())
			snapshot := *v
			req.RawRequest.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return ioutil.NopCloser(&r), nil
			}
		default:
			// This is where we'd set it to -1 (at least
			// if body != NoBody) to mean unknown, but
			// that broke people during the Go 1.8 testing
			// period. People depend on it being 0 I
			// guess. Maybe retry later. See Issue 18117.
		}
		// For client requests, Request.ContentLength of 0
		// means either actually 0, or unknown. The only way
		// to explicitly say that the ContentLength is zero is
		// to set the Body to nil. But turns out too much code
		// depends on NewRequest returning a non-nil Body,
		// so we use a well-known ReadCloser variable instead
		// and have the http package also treat that sentinel
		// variable to mean explicitly zero.
		if req.RawRequest.GetBody != nil && req.RawRequest.ContentLength == 0 {
			req.RawRequest.Body = http.NoBody
			req.RawRequest.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}

	return req
}

// SetHost sets host for the HTTP request.
func (req *Request) SetHost(host string) *Request {
	if req.Err != nil {
		return req
	}

	req.Host = host
	return req
}

// SetHeaders sets headers for the HTTP request.
func (req *Request) SetHeaders(headers Headers) *Request {
	if req.Err != nil {
		return req
	}

	for k, v := range headers {
		req.Headers.Set(k, v)
	}
	return req
}

// SetContentType sets Content-Type header value for the HTTP request.
func (req *Request) SetContentType(contentType string) *Request {
	if req.Err != nil {
		return req
	}

	req.Headers.Set("Content-Type", contentType)
	return req
}

// SetUserAgent sets User-Agent header value for the HTTP request.
func (req *Request) SetUserAgent(userAgent string) *Request {
	if req.Err != nil {
		return req
	}

	req.Headers.Set("User-Agent", userAgent)
	return req
}

// SetReferer sets Referer header value for the HTTP request.
func (req *Request) SetReferer(referer string) *Request {
	if req.Err != nil {
		return req
	}

	req.Headers.Set("Referer", referer)
	return req
}

// SetQuery sets query params for the HTTP request.
func (req *Request) SetQuery(params Params) *Request {
	if req.Err != nil {
		return req
	}

	query := req.RawRequest.URL.Query()
	for k, v := range params {
		query.Add(k, v)
	}

	req.RawRequest.URL.RawQuery = query.Encode()
	return req
}

// SetContent sets bytes payload for the HTTP request.
func (req *Request) SetContent(content []byte) *Request {
	if req.Err != nil {
		return req
	}

	r := bytes.NewBuffer(content)
	req.SetBody(r)
	return req
}

// SetText sets plain text payload for the HTTP request.
func (req *Request) SetText(text string) *Request {
	if req.Err != nil {
		return req
	}

	r := bytes.NewBufferString(text)
	req.SetBody(r)
	req.SetContentType("text/plain; charset=utf-8")
	return req
}

// SetForm sets form payload for the HTTP request.
func (req *Request) SetForm(form Form) *Request {
	if req.Err != nil {
		return req
	}

	data := stdurl.Values{}
	for k, v := range form {
		data.Add(k, v)
	}

	r := strings.NewReader(data.Encode())
	req.SetBody(r)
	req.SetContentType("application/x-www-form-urlencoded")
	return req
}

// SetJSON sets json payload for the HTTP request.
func (req *Request) SetJSON(data JSON, escapeHTML bool) *Request {
	if req.Err != nil {
		return req
	}

	b, err := jsonMarshal(data, "", "", escapeHTML)
	if err != nil {
		req.raiseError("SetJSON", err)
		return req
	}

	r := bytes.NewReader(b)
	req.SetBody(r)
	req.SetContentType("application/json")
	return req
}

// SetFiles sets files payload for the HTTP request.
func (req *Request) SetFiles(files Files) *Request {
	if req.Err != nil {
		return req
	}

	for fieldName, filePath := range files {
		if _, err := existsFile(filePath); err != nil {
			req.raiseError("SetFiles",
				fmt.Errorf("file for [%s] not ready: %s", fieldName, err.Error()))
			return req
		}
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer mw.Close()

		for fieldName, filePath := range files {
			fileName := filepath.Base(filePath)
			part, err := mw.CreateFormFile(fieldName, fileName)
			if err != nil {
				return
			}

			file, err := os.Open(filePath)
			if err != nil {
				return
			}

			_, err = io.Copy(part, file)
			if err != nil || file.Close() != nil {
				return
			}
		}
	}()

	req.SetBody(pr)
	req.SetContentType(mw.FormDataContentType())
	return req
}

func existsFile(filename string) (bool, error) {
	fi, err := os.Stat(filename)
	if err == nil {
		if fi.Mode().IsDir() {
			return false, fmt.Errorf("%q is a directory", filename)
		}
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, err
	}

	return true, err
}

// SetCookies sets cookies for the HTTP request.
func (req *Request) SetCookies(cookies ...*http.Cookie) *Request {
	if req.Err != nil {
		return req
	}

	req.Cookies = append(req.Cookies, cookies...)
	return req
}

// SetBasicAuth sets basic authentication for the HTTP request.
func (req *Request) SetBasicAuth(username string, password string) *Request {
	if req.Err != nil {
		return req
	}

	req.auth = &auth{
		username: username,
		password: password,
	}
	return req
}

// SetBearerToken sets bearer token for the HTTP request.
func (req *Request) SetBearerToken(token string) *Request {
	if req.Err != nil {
		return req
	}

	req.bearerToken = token
	return req
}

// SetContext sets context for the HTTP request.
func (req *Request) SetContext(ctx context.Context) *Request {
	if req.Err != nil {
		return req
	}

	if ctx == nil {
		req.raiseError("SetContext", ErrNilContext)
		return req
	}

	req.ctx = ctx
	return req
}

// SetTimeout sets timeout for the HTTP request.
func (req *Request) SetTimeout(timeout time.Duration) *Request {
	if req.Err != nil {
		return req
	}

	req.Timeout = timeout
	return req
}

// SetRetry sets retry policy for the HTTP request.
// Notes: Request timeout or context has priority over the retry policy.
func (req *Request) SetRetry(attempts int, delay time.Duration,
	conditions ...func(*Response) bool) *Request {
	if req.Err != nil {
		return req
	}

	if attempts > 1 {
		req.retry = &retry{
			attempts:   attempts,
			delay:      delay,
			conditions: conditions,
		}
	}
	return req
}

// WithBody sets body for the HTTP request.
func WithBody(body io.Reader) RequestOption {
	return func(req *Request) *Request {
		return req.SetBody(body)
	}
}

// WithHost sets host for the HTTP request.
func WithHost(host string) RequestOption {
	return func(req *Request) *Request {
		return req.SetHost(host)
	}
}

// WithHeaders sets headers for the HTTP request.
func WithHeaders(headers Headers) RequestOption {
	return func(req *Request) *Request {
		return req.SetHeaders(headers)
	}
}

// WithContentType sets contentType header value for the HTTP request.
func WithContentType(contentType string) RequestOption {
	return func(req *Request) *Request {
		return req.SetContentType(contentType)
	}
}

// WithUserAgent sets User-Agent header value for the HTTP request.
func WithUserAgent(userAgent string) RequestOption {
	return func(req *Request) *Request {
		return req.SetUserAgent(userAgent)
	}
}

// WithReferer sets Referer header value for the HTTP request.
func WithReferer(referer string) RequestOption {
	return func(req *Request) *Request {
		return req.SetReferer(referer)
	}
}

// WithQuery sets query params for the HTTP request.
func WithQuery(params Params) RequestOption {
	return func(req *Request) *Request {
		return req.SetQuery(params)
	}
}

// WithContent sets bytes payload for the HTTP request.
func WithContent(content []byte) RequestOption {
	return func(req *Request) *Request {
		return req.SetContent(content)
	}
}

// WithText sets plain text payload for the HTTP request.
func WithText(text string) RequestOption {
	return func(req *Request) *Request {
		return req.SetText(text)
	}
}

// WithForm sets form payload for the HTTP request.
func WithForm(form Form) RequestOption {
	return func(req *Request) *Request {
		return req.SetForm(form)
	}
}

// WithJSON sets json payload for the HTTP request.
func WithJSON(data JSON, escapeHTML bool) RequestOption {
	return func(req *Request) *Request {
		return req.SetJSON(data, escapeHTML)
	}
}

// WithFiles sets files payload for the HTTP request.
func WithFiles(files Files) RequestOption {
	return func(req *Request) *Request {
		return req.SetFiles(files)
	}
}

// WithCookies appends cookies for the HTTP request.
func WithCookies(cookies ...*http.Cookie) RequestOption {
	return func(req *Request) *Request {
		return req.SetCookies(cookies...)
	}
}

// WithBasicAuth sets basic authentication for the HTTP request.
func WithBasicAuth(username string, password string) RequestOption {
	return func(req *Request) *Request {
		return req.SetBasicAuth(username, password)
	}
}

// WithBearerToken sets bearer token for the HTTP request.
func WithBearerToken(token string) RequestOption {
	return func(req *Request) *Request {
		return req.SetBearerToken(token)
	}
}

// WithContext sets context for the HTTP request.
func WithContext(ctx context.Context) RequestOption {
	return func(req *Request) *Request {
		return req.SetContext(ctx)
	}
}

// WithTimeout sets timeout for the HTTP request.
func WithTimeout(timeout time.Duration) RequestOption {
	return func(req *Request) *Request {
		return req.SetTimeout(timeout)
	}
}

// WithRetry sets retry policy for the HTTP request.
// Notes: Request timeout or context has priority over the retry policy.
func WithRetry(attempts int, delay time.Duration,
	conditions ...func(*Response) bool) RequestOption {
	return func(req *Request) *Request {
		return req.SetRetry(attempts, delay, conditions...)
	}
}
