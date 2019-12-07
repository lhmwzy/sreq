package sreq

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	stdurl "net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	// DefaultTimeout is the timeout used by DefaultClient.
	DefaultTimeout = 120 * time.Second
)

var (
	// DefaultClient is the default sreq Client.
	DefaultClient = New()
)

type (
	// Client wraps the raw HTTP client.
	// Do not modify the client across Goroutines!
	// You should reuse it once initialized.
	Client struct {
		RawClient *http.Client
		Host      string
		Headers   Headers
		Cookies   []*http.Cookie
		Err       error

		auth        *auth
		bearerToken string
		ctx         context.Context
		retry       *retry

		requestInterceptors  []RequestInterceptor
		responseInterceptors []ResponseInterceptor
	}
)

// New returns a new Client.
// It's a clone of DefaultClient indeed.
func New() *Client {
	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	rawClient := &http.Client{
		Transport: DefaultTransport(),
		Jar:       jar,
		Timeout:   DefaultTimeout,
	}
	client := &Client{
		RawClient: rawClient,
		Headers:   make(Headers),
	}
	return client
}

func (c *Client) httpTransport() (*http.Transport, error) {
	t, ok := c.RawClient.Transport.(*http.Transport)
	if !ok || t == nil {
		return nil, ErrUnexpectedTransport
	}

	return t, nil
}

func (c *Client) raiseError(cause string, err error) {
	c.Err = &ClientError{
		Cause: cause,
		Err:   err,
	}
}

// Raw returns the raw HTTP client.
func (c *Client) Raw() (*http.Client, error) {
	return c.RawClient, c.Err
}

// SetTransport sets transport of the HTTP client.
func SetTransport(transport http.RoundTripper) *Client {
	return DefaultClient.SetTransport(transport)
}

// SetTransport sets transport of the HTTP client.
func (c *Client) SetTransport(transport http.RoundTripper) *Client {
	if c.Err != nil {
		return c
	}

	c.RawClient.Transport = transport
	return c
}

// SetRedirect sets policy of the HTTP client for handling redirects.
func SetRedirect(policy func(req *http.Request, via []*http.Request) error) *Client {
	return DefaultClient.SetRedirect(policy)
}

// SetRedirect sets policy of the HTTP client for handling redirects.
func (c *Client) SetRedirect(policy func(req *http.Request, via []*http.Request) error) *Client {
	if c.Err != nil {
		return c
	}

	c.RawClient.CheckRedirect = policy
	return c
}

// DisableRedirect makes the HTTP client not follow redirects.
func DisableRedirect() *Client {
	return DefaultClient.DisableRedirect()
}

// DisableRedirect makes the HTTP client not follow redirects.
func (c *Client) DisableRedirect() *Client {
	policy := func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c.SetRedirect(policy)
}

// SetCookieJar sets cookie jar of the HTTP client.
func SetCookieJar(jar http.CookieJar) *Client {
	return DefaultClient.SetCookieJar(jar)
}

// SetCookieJar sets cookie jar of the HTTP client.
func (c *Client) SetCookieJar(jar http.CookieJar) *Client {
	if c.Err != nil {
		return c
	}

	c.RawClient.Jar = jar
	return c
}

// DisableSession makes the HTTP client not use cookie jar.
// Only use if you don't want to keep session for the next HTTP request.
func DisableSession() *Client {
	return DefaultClient.DisableSession()
}

// DisableSession makes the HTTP client not use cookie jar.
// Only use if you don't want to keep session for the next HTTP request.
func (c *Client) DisableSession() *Client {
	return c.SetCookieJar(nil)
}

// SetTimeout sets timeout of the HTTP client.
func SetTimeout(timeout time.Duration) *Client {
	return DefaultClient.SetTimeout(timeout)
}

// SetTimeout sets timeout of the HTTP client.
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	if c.Err != nil {
		return c
	}

	c.RawClient.Timeout = timeout
	return c
}

// SetProxy sets proxy of the HTTP client.
func SetProxy(proxy func(*http.Request) (*stdurl.URL, error)) *Client {
	return DefaultClient.SetProxy(proxy)
}

// SetProxy sets proxy of the HTTP client.
func (c *Client) SetProxy(proxy func(*http.Request) (*stdurl.URL, error)) *Client {
	if c.Err != nil {
		return c
	}

	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("SetProxy", err)
		return c
	}

	t.Proxy = proxy
	c.RawClient.Transport = t
	return c
}

// SetProxyFromURL sets proxy of the HTTP client from a url.
func SetProxyFromURL(url string) *Client {
	return DefaultClient.SetProxyFromURL(url)
}

// SetProxyFromURL sets proxy of the HTTP client from a url.
func (c *Client) SetProxyFromURL(url string) *Client {
	if c.Err != nil {
		return c
	}

	fixedURL, err := stdurl.Parse(url)
	if err != nil {
		c.raiseError("SetProxyFromURL", err)
		return c
	}
	return c.SetProxy(http.ProxyURL(fixedURL))
}

// DisableProxy makes the HTTP client not use proxy.
func DisableProxy() *Client {
	return DefaultClient.DisableProxy()
}

// DisableProxy makes the HTTP client not use proxy.
func (c *Client) DisableProxy() *Client {
	return c.SetProxy(nil)
}

// SetTLSClientConfig sets TLS configuration of the HTTP client.
func SetTLSClientConfig(config *tls.Config) *Client {
	return DefaultClient.SetTLSClientConfig(config)
}

// SetTLSClientConfig sets TLS configuration of the HTTP client.
func (c *Client) SetTLSClientConfig(config *tls.Config) *Client {
	if c.Err != nil {
		return c
	}

	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("SetTLSClientConfig", err)
		return c
	}

	t.TLSClientConfig = config
	c.RawClient.Transport = t
	return c
}

// AppendClientCertificates appends client certificates to the HTTP client.
func AppendClientCertificates(certs ...tls.Certificate) *Client {
	return DefaultClient.AppendClientCertificates(certs...)
}

// AppendClientCertificates appends client certificates to the HTTP client.
func (c *Client) AppendClientCertificates(certs ...tls.Certificate) *Client {
	if c.Err != nil {
		return c
	}

	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("AppendClientCertificates", err)
		return c
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}

	t.TLSClientConfig.Certificates = append(t.TLSClientConfig.Certificates, certs...)
	c.RawClient.Transport = t
	return c
}

// AppendRootCAs appends root certificate authorities to the HTTP client.
func AppendRootCAs(pemFilePath string) *Client {
	return DefaultClient.AppendRootCAs(pemFilePath)
}

// AppendRootCAs appends root certificate authorities to the HTTP client.
func (c *Client) AppendRootCAs(pemFilePath string) *Client {
	if c.Err != nil {
		return c
	}

	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("AppendRootCAs", err)
		return c
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
	if t.TLSClientConfig.RootCAs == nil {
		t.TLSClientConfig.RootCAs = x509.NewCertPool()
	}

	pemCerts, err := ioutil.ReadFile(pemFilePath)
	if err != nil {
		c.raiseError("AppendRootCAs", err)
		return c
	}

	t.TLSClientConfig.RootCAs.AppendCertsFromPEM(pemCerts)
	c.RawClient.Transport = t
	return c
}

// DisableVerify makes the HTTP client not verify the server's TLS certificate.
func DisableVerify() *Client {
	return DefaultClient.DisableVerify()
}

// DisableVerify makes the HTTP client not verify the server's TLS certificate.
func (c *Client) DisableVerify() *Client {
	if c.Err != nil {
		return c
	}

	t, err := c.httpTransport()
	if err != nil {
		c.raiseError("DisableVerify", err)
		return c
	}

	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}

	t.TLSClientConfig.InsecureSkipVerify = true
	c.RawClient.Transport = t
	return c
}

// SetHost sets host of the client.
// The host will be applied to all requests raised from this client instance.
// Also it can be overridden at request level host options.
func SetHost(host string) *Client {
	return DefaultClient.SetHost(host)
}

// SetHost sets host of the client.
// The host will be applied to all requests raised from this client instance.
// Also it can be overridden at request level host options.
func (c *Client) SetHost(host string) *Client {
	if c.Err != nil {
		return c
	}

	c.Host = host
	return c
}

// SetHeaders sets headers of the client.
// These headers will be applied to all requests raised from this client instance.
// Also it can be overridden at request level headers options.
func SetHeaders(headers Headers) *Client {
	return DefaultClient.SetHeaders(headers)
}

// SetHeaders sets headers of the client.
// These headers will be applied to all requests raised from this client instance.
// Also it can be overridden at request level headers options.
func (c *Client) SetHeaders(headers Headers) *Client {
	if c.Err != nil {
		return c
	}

	for k, v := range headers {
		c.Headers.Set(k, v)
	}
	return c
}

// SetUserAgent sets User-Agent header value of the client.
func SetUserAgent(userAgent string) *Client {
	return DefaultClient.SetUserAgent(userAgent)
}

// SetUserAgent sets User-Agent header value of the client.
func (c *Client) SetUserAgent(userAgent string) *Client {
	if c.Err != nil {
		return c
	}

	c.Headers.Set("User-Agent", userAgent)
	return c
}

// SetReferer sets Referer header value of the client.
func SetReferer(referer string) *Client {
	return DefaultClient.SetReferer(referer)
}

// SetReferer sets Referer header value of the client.
func (c *Client) SetReferer(referer string) *Client {
	if c.Err != nil {
		return c
	}

	c.Headers.Set("Referer", referer)
	return c
}

// SetCookies sets cookies of the client.
// These cookies will be applied to all requests raised from this client instance.
func SetCookies(cookies ...*http.Cookie) *Client {
	return DefaultClient.SetCookies(cookies...)
}

// SetCookies sets cookies of the client.
// These cookies will be applied to all requests raised from this client instance.
func (c *Client) SetCookies(cookies ...*http.Cookie) *Client {
	if c.Err != nil {
		return c
	}

	c.Cookies = append(c.Cookies, cookies...)
	return c
}

// SetBasicAuth sets basic authentication of the client.
// The basic authentication will be applied to all requests raised from this client instance.
// Also it can be overridden at request level basic authentication options.
func SetBasicAuth(username string, password string) *Client {
	return DefaultClient.SetBasicAuth(username, password)
}

// SetBasicAuth sets basic authentication of the client.
// The basic authentication will be applied to all requests raised from this client instance.
// Also it can be overridden at request level basic authentication options.
func (c *Client) SetBasicAuth(username string, password string) *Client {
	if c.Err != nil {
		return c
	}

	c.auth = &auth{
		username: username,
		password: password,
	}
	return c
}

// SetBearerToken sets bearer token of the client.
// The bearer token will be applied to all requests raised from this client instance.
// Also it can be overridden at request level bearer token options.
func SetBearerToken(token string) *Client {
	return DefaultClient.SetBearerToken(token)
}

// SetBearerToken sets bearer token of the client.
// The bearer token will be applied to all requests raised from this client instance.
// Also it can be overridden at request level bearer token options.
func (c *Client) SetBearerToken(token string) *Client {
	if c.Err != nil {
		return c
	}

	c.bearerToken = token
	return c
}

// SetContext sets context of the client.
// The context will be applied to all requests raised from this client instance.
// Also it can be overridden at request level context options.
func SetContext(ctx context.Context) *Client {
	return DefaultClient.SetContext(ctx)
}

// SetContext sets context of the client.
// The context will be applied to all requests raised from this client instance.
// Also it can be overridden at request level context options.
func (c *Client) SetContext(ctx context.Context) *Client {
	if c.Err != nil {
		return c
	}

	if ctx == nil {
		c.raiseError("SetContext", ErrNilContext)
		return c
	}

	c.ctx = ctx
	return c
}

// SetRetry sets retry policy of the client.
// The retry policy will be applied to all requests raised from this client instance.
// Also it can be overridden at request level retry policy options.
// Notes: Request timeout or context has priority over the retry policy.
func SetRetry(attempts int, delay time.Duration,
	conditions ...func(*Response) bool) *Client {
	return DefaultClient.SetRetry(attempts, delay, conditions...)
}

// SetRetry sets retry policy of the client.
// The retry policy will be applied to all requests raised from this client instance.
// Also it can be overridden at request level retry policy options.
// Notes: Request timeout or context has priority over the retry policy.
func (c *Client) SetRetry(attempts int, delay time.Duration,
	conditions ...func(*Response) bool) *Client {
	if c.Err != nil {
		return c
	}

	if attempts > 1 {
		c.retry = &retry{
			attempts:   attempts,
			delay:      delay,
			conditions: conditions,
		}
	}
	return c
}

// UseRequestInterceptors appends request interceptors of the client.
func UseRequestInterceptors(interceptors ...RequestInterceptor) *Client {
	return DefaultClient.UseRequestInterceptors(interceptors...)
}

// UseRequestInterceptors appends request interceptors of the client.
func (c *Client) UseRequestInterceptors(interceptors ...RequestInterceptor) *Client {
	if c.Err != nil {
		return c
	}

	c.requestInterceptors = append(c.requestInterceptors, interceptors...)
	return c
}

// UseResponseInterceptors appends response interceptors of the client.
func UseResponseInterceptors(interceptors ...ResponseInterceptor) *Client {
	return DefaultClient.UseResponseInterceptors(interceptors...)
}

// UseResponseInterceptors appends response interceptors of the client.
func (c *Client) UseResponseInterceptors(interceptors ...ResponseInterceptor) *Client {
	if c.Err != nil {
		return c
	}

	c.responseInterceptors = append(c.responseInterceptors, interceptors...)
	return c
}

// Get makes a GET HTTP request.
func Get(url string, opts ...RequestOption) *Response {
	return DefaultClient.Get(url, opts...)
}

// Get makes a GET HTTP request.
func (c *Client) Get(url string, opts ...RequestOption) *Response {
	return c.Send(MethodGet, url, opts...)
}

// Head makes a HEAD HTTP request.
func Head(url string, opts ...RequestOption) *Response {
	return DefaultClient.Head(url, opts...)
}

// Head makes a HEAD HTTP request.
func (c *Client) Head(url string, opts ...RequestOption) *Response {
	return c.Send(MethodHead, url, opts...)
}

// Post makes a POST HTTP request.
func Post(url string, opts ...RequestOption) *Response {
	return DefaultClient.Post(url, opts...)
}

// Post makes a POST HTTP request.
func (c *Client) Post(url string, opts ...RequestOption) *Response {
	return c.Send(MethodPost, url, opts...)
}

// Put makes a PUT HTTP request.
func Put(url string, opts ...RequestOption) *Response {
	return DefaultClient.Put(url, opts...)
}

// Put makes a PUT HTTP request.
func (c *Client) Put(url string, opts ...RequestOption) *Response {
	return DefaultClient.Send(MethodPut, url, opts...)
}

// Patch makes a PATCH HTTP request.
func Patch(url string, opts ...RequestOption) *Response {
	return DefaultClient.Patch(url, opts...)
}

// Patch makes a PATCH HTTP request.
func (c *Client) Patch(url string, opts ...RequestOption) *Response {
	return c.Send(MethodPatch, url, opts...)
}

// Delete makes a DELETE HTTP request.
func Delete(url string, opts ...RequestOption) *Response {
	return DefaultClient.Delete(url, opts...)
}

// Delete makes a DELETE HTTP request.
func (c *Client) Delete(url string, opts ...RequestOption) *Response {
	return c.Send(MethodDelete, url, opts...)
}

// Send makes an HTTP request using a specified method.
func Send(method string, url string, opts ...RequestOption) *Response {
	return DefaultClient.Send(method, url, opts...)
}

// Send makes an HTTP request using a specified method.
func (c *Client) Send(method string, url string, opts ...RequestOption) *Response {
	req := NewRequest(method, url)
	for _, opt := range opts {
		req = opt(req)
	}
	return c.Do(req)
}

// FilterCookies returns the cookies to send in a request for the given URL.
func FilterCookies(url string) ([]*http.Cookie, error) {
	return DefaultClient.FilterCookies(url)
}

// FilterCookies returns the cookies to send in a request for the given URL.
func (c *Client) FilterCookies(url string) ([]*http.Cookie, error) {
	if c.RawClient.Jar == nil {
		return nil, ErrNilCookieJar
	}

	u, err := stdurl.Parse(url)
	if err != nil {
		return nil, err
	}
	cookies := c.RawClient.Jar.Cookies(u)
	if len(cookies) == 0 {
		return nil, ErrJarCookiesNotPresent
	}

	return cookies, nil
}

// FilterCookie returns the named cookie to send in a request for the given URL.
func FilterCookie(url string, name string) (*http.Cookie, error) {
	return DefaultClient.FilterCookie(url, name)
}

// FilterCookie returns the named cookie to send in a request for the given URL.
func (c *Client) FilterCookie(url string, name string) (*http.Cookie, error) {
	cookies, err := c.FilterCookies(url)
	if err != nil {
		return nil, err
	}

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie, nil
		}
	}

	return nil, ErrJarNamedCookieNotPresent
}

// Do sends a request and returns its response.
func Do(req *Request) *Response {
	return DefaultClient.Do(req)
}

// Do sends a request and returns its  response.
func (c *Client) Do(req *Request) *Response {
	resp := new(Response)

	if c.Err != nil {
		resp.Err = c.Err
		return resp
	}

	if req.Err != nil {
		resp.Err = req.Err
		return resp
	}

	c.setHost(req)
	c.setHeaders(req)
	c.setCookies(req)
	c.setBasicAuth(req)
	c.setBearerToken(req)
	c.setContext(req)

	err := c.onBeforeRequest(req)
	if err != nil {
		resp.Err = err
		return resp
	}

	c.doWithRetry(req, resp)
	c.onAfterResponse(resp)
	return resp
}

func (c *Client) onBeforeRequest(req *Request) error {
	var err error
	for _, interceptor := range c.requestInterceptors {
		if err = interceptor(req); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) onAfterResponse(resp *Response) {
	var err error
	for _, interceptor := range c.responseInterceptors {
		if err = interceptor(resp); err != nil {
			resp.Err = err
			return
		}
	}
}

func (c *Client) setHost(req *Request) {
	host := c.Host
	if req.Host != "" {
		host = req.Host
	}

	if host != "" {
		req.RawRequest.Host = host
	}
}

func (c *Client) setHeaders(req *Request) {
	headers := make(Headers, len(c.Headers))
	for k, v := range c.Headers {
		headers.Set(k, v)
	}
	for k, v := range req.Headers {
		headers.Set(k, v)
	}

	for k, v := range headers {
		switch v := v.(type) {
		case string:
			req.RawRequest.Header.Set(k, v)
		case int:
			req.RawRequest.Header.Set(k, strconv.Itoa(v))
		case []string:
			for _, vv := range v {
				req.RawRequest.Header.Add(k, vv)
			}
		case []int:
			for _, vv := range v {
				req.RawRequest.Header.Add(k, strconv.Itoa(vv))
			}
		case []interface{}:
			for _, vv := range v {
				switch vv := vv.(type) {
				case string:
					req.RawRequest.Header.Add(k, vv)
				case int:
					req.RawRequest.Header.Add(k, strconv.Itoa(vv))
				}
			}
		}
	}
}

func (c *Client) setCookies(req *Request) {
	for _, c := range c.Cookies {
		req.RawRequest.AddCookie(c)
	}

	for _, c := range req.Cookies {
		req.RawRequest.AddCookie(c)
	}
}

func (c *Client) setBasicAuth(req *Request) {
	auth := c.auth
	if req.auth != nil {
		auth = req.auth
	}

	if auth != nil {
		req.RawRequest.Header.Set("Authorization", "Basic "+basicAuth(auth.username, auth.password))
	}
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *Client) setBearerToken(req *Request) {
	token := c.bearerToken
	if req.bearerToken != "" {
		token = req.bearerToken
	}

	if token != "" {
		req.RawRequest.Header.Set("Authorization", "Bearer "+token)
	}
}

func (c *Client) setContext(req *Request) {
	ctx := c.ctx
	if req.ctx != nil {
		ctx = req.ctx
	}

	if ctx != nil {
		req.RawRequest = req.RawRequest.WithContext(ctx)
	}
}

func (c *Client) doWithRetry(req *Request, resp *Response) {
	ctx := req.RawRequest.Context()
	var cancel context.CancelFunc
	if req.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		req.RawRequest = req.RawRequest.WithContext(ctx)
		defer cancel()
	}

	if req.retry == nil && c.retry == nil {
		resp.RawResponse, resp.Err = c.do(req.RawRequest)
		return
	}

	retry := req.retry
	if retry == nil {
		retry = c.retry
	}

	var err error
	for i := retry.attempts; i > 0; i-- {
		resp.RawResponse, resp.Err = c.do(req.RawRequest)
		if err = ctx.Err(); err != nil {
			resp.Err = err
			return
		}

		shouldRetry := resp.Err != nil
		for _, condition := range retry.conditions {
			shouldRetry = condition(resp)
			if shouldRetry {
				break
			}
		}

		if !shouldRetry {
			return
		}

		select {
		case <-time.After(retry.delay):
		case <-ctx.Done():
			resp.Err = ctx.Err()
			return
		}
	}
}

func (c *Client) do(rawRequest *http.Request) (*http.Response, error) {
	rawResponse, err := c.RawClient.Do(rawRequest)
	if err != nil {
		return rawResponse, err
	}

	if strings.EqualFold(rawResponse.Header.Get("Content-Encoding"), "gzip") &&
		rawResponse.ContentLength != 0 {
		if _, ok := rawResponse.Body.(*gzip.Reader); !ok {
			rawResponse.Body, err = gzip.NewReader(rawResponse.Body)
			return rawResponse, err
		}
	}

	return rawResponse, nil
}
