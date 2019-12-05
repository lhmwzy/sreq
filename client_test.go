package sreq_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/winterssy/sreq"
	"golang.org/x/net/publicsuffix"
)

func TestClient_RaiseError(t *testing.T) {
	const (
		invalidURL   = "http://127.0.0.1:1081^"
		validURL     = "http://127.0.0.1:1081"
		pemFileExist = "./testdata/root-ca.pem"
	)

	_, err := sreq.New().
		SetProxyFromURL(invalidURL).
		SetTransport(nil).
		DisableSession().
		DisableRedirect().
		SetTimeout(120*time.Second).
		SetProxyFromURL(validURL).
		SetProxy(http.ProxyFromEnvironment).
		SetTLSClientConfig(&tls.Config{}).
		AppendClientCertificates(tls.Certificate{}).
		AppendRootCAs(pemFileExist).
		DisableVerify().
		SetHost("httpbin.org").
		SetHeaders(sreq.Headers{
			"Origin": "http://httpbin.org",
		}).
		SetUserAgent("Go-http-client").
		SetReferer("http://httpbin.org").
		SetCookies(
			&http.Cookie{
				Name:  "uid",
				Value: "10086",
			},
		).
		SetBasicAuth("user", "pass").
		SetBearerToken("sreq").
		SetContext(context.Background()).
		SetRetry(3, 1*time.Second).
		UseRequestInterceptors(func(req *sreq.Request) error {
			return nil
		}).
		UseResponseInterceptors(func(resp *sreq.Response) error {
			return nil
		}).
		Raw()
	if err == nil {
		t.Error("Client_RaiseError test failed")
	}
}

func TestClient_DisableSession(t *testing.T) {
	rawClient, err := sreq.New().DisableSession().Raw()
	if err != nil {
		t.Fatal(err)
	}
	if rawClient.Jar != nil {
		t.Error("Client_DisableSession test failed")
	}
}

func TestClient_SetTimeout(t *testing.T) {
	const (
		timeout = 3 * time.Second
	)

	rawClient, err := sreq.New().SetTimeout(timeout).Raw()
	if err != nil {
		t.Fatal(err)
	}
	if got := rawClient.Timeout; got != timeout {
		t.Error("Client_SetTimeout test failed")
	}
}

func TestClient_DisableRedirect(t *testing.T) {
	rawClient, err := sreq.New().DisableRedirect().Raw()
	if err != nil {
		t.Fatal(err)
	}
	if err = rawClient.CheckRedirect(nil, nil); err != http.ErrUseLastResponse {
		t.Error("Client_DisableRedirect test failed")
	}
}

func TestClient_SetProxyFromURL(t *testing.T) {
	const (
		url        = "http://127.0.0.1:1081"
		invalidURL = "http://127.0.0.1:1081^"
	)

	_, err := sreq.New().SetProxyFromURL(invalidURL).Raw()
	if err == nil {
		t.Error("Client_SetProxyFromURL test failed")
	}

	rawClient, err := sreq.New().SetProxyFromURL(url).Raw()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := rawClient.Transport.(*http.Transport)
	if !ok || transport == nil || transport.Proxy == nil {
		t.Fatal("Client_SetProxyFromURL test failed")
	}

	req, _ := http.NewRequest("GET", "https://www.google.com", nil)
	proxyURL, err := transport.Proxy(req)
	if err != nil {
		t.Fatal(err)
	}
	if got := proxyURL.String(); got != url {
		t.Error("Client_SetProxyFromURL test failed")
	}

	rawClient, err = sreq.New().SetTransport(nil).SetProxyFromURL(url).Raw()
	if err == nil {
		t.Error("Client_SetProxyFromURL test failed")
	}
}

func TestClient_DisableProxy(t *testing.T) {
	rawClient, err := sreq.New().DisableProxy().Raw()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := rawClient.Transport.(*http.Transport)
	if !ok || transport == nil || transport.Proxy != nil {
		t.Error("Client_DisableProxy test failed")
	}
}

func TestClient_SetTLSClientConfig(t *testing.T) {
	config := &tls.Config{}

	rawClient, err := sreq.New().SetTLSClientConfig(config).Raw()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := rawClient.Transport.(*http.Transport)
	if !ok || transport == nil || transport.TLSClientConfig == nil {
		t.Error("Client_SetTLSClientConfig test failed")
	}

	rawClient, err = sreq.New().SetTransport(nil).SetTLSClientConfig(config).Raw()
	if err == nil {
		t.Error("Client_SetTLSClientConfig test failed")
	}
}

func TestClient_AppendClientCertificates(t *testing.T) {
	cert := tls.Certificate{}

	rawClient, err := sreq.New().AppendClientCertificates(cert).Raw()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := rawClient.Transport.(*http.Transport)
	if !ok || transport == nil || transport.TLSClientConfig == nil ||
		len(transport.TLSClientConfig.Certificates) != 1 {
		t.Fatal("Client_AppendClientCertificates test failed")
	}

	rawClient, err = sreq.New().SetTransport(nil).AppendClientCertificates(cert).Raw()
	if err == nil {
		t.Error("Client_AppendClientCertificates test failed")
	}
}

func TestClient_AppendRootCAs(t *testing.T) {
	const (
		pemFileExist    = "./testdata/root-ca.pem"
		pemFileNotExist = "./testdata/root-ca-not-exist.pem"
	)

	rawClient, err := sreq.New().AppendRootCAs(pemFileExist).Raw()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := rawClient.Transport.(*http.Transport)
	if !ok || transport == nil || transport.TLSClientConfig == nil ||
		transport.TLSClientConfig.RootCAs == nil {
		t.Error("Client_AppendRootCAs test failed")
	}

	rawClient, err = sreq.New().AppendRootCAs(pemFileNotExist).Raw()
	if err == nil {
		t.Error("Client_AppendRootCAs test failed")
	}

	rawClient, err = sreq.New().SetTransport(nil).AppendRootCAs(pemFileExist).Raw()
	if err == nil {
		t.Error("Client_AppendRootCAs test failed")
	}
}

func TestClient_DisableVerify(t *testing.T) {
	rawClient, err := sreq.New().SetTransport(nil).DisableVerify().Raw()
	if err == nil {
		t.Error("Client_DisableVerify test failed")
	}

	rawClient, err = sreq.New().DisableVerify().Raw()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := rawClient.Transport.(*http.Transport)
	if !ok || transport == nil || transport.TLSClientConfig == nil ||
		!transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("Client_DisableVerify test failed")
	}
}

func TestClient_SetHost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Host))
	}))
	defer ts.Close()

	client := sreq.New().SetHost("github.com")
	data, err := client.
		Get(ts.URL).
		EnsureStatusOk().
		Text()
	if err != nil {
		t.Fatal(err)
	}

	if data != "github.com" {
		t.Error("Client_SetHost test failed")
	}

	data, err = client.
		Get(ts.URL,
			sreq.WithHost("google.com"),
		).
		EnsureStatusOk().
		Text()
	if err != nil {
		t.Fatal(err)
	}

	if data != "google.com" {
		t.Error("Client_SetHost test failed")
	}
}

func TestClient_SetHeaders(t *testing.T) {
	type response struct {
		Headers map[string]string `json:"headers"`
	}

	client := sreq.New().SetHeaders(sreq.Headers{
		"Origin":  "http://httpbin.org",
		"Referer": "https://www.google.com",
	})
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get",
			sreq.WithHeaders(sreq.Headers{
				"Referer": "http://httpbin.org",
			}),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Headers["Origin"] != "http://httpbin.org" || resp.Headers["Referer"] != "http://httpbin.org" {
		t.Error("Client_SetHeaders test failed")
	}
}

func TestClient_SetUserAgent(t *testing.T) {
	type response struct {
		Headers map[string]string `json:"headers"`
	}

	client := sreq.New().SetUserAgent("Go-http-client")
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get").
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Headers["User-Agent"] != "Go-http-client" {
		t.Error("Client_SetUserAgent test failed")
	}

	_resp := new(response)
	err = client.
		Get("http://httpbin.org/get",
			sreq.WithUserAgent("Go-sreq"),
		).
		EnsureStatusOk().
		JSON(_resp)
	if err != nil {
		t.Fatal(err)
	}

	if _resp.Headers["User-Agent"] != "Go-sreq" {
		t.Error("Client_SetUserAgent test failed")
	}
}

func TestClient_SetReferer(t *testing.T) {
	type response struct {
		Headers map[string]string `json:"headers"`
	}

	client := sreq.New().SetReferer("https://www.google.com")
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get").
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Headers["Referer"] != "https://www.google.com" {
		t.Error("Client_SetReferer test failed")
	}
}

func TestClient_SetCookies(t *testing.T) {
	type response struct {
		Cookies map[string]string `json:"cookies"`
	}

	client := sreq.New().SetCookies(&http.Cookie{
		Name:  "n1",
		Value: "v1",
	})
	resp := new(response)
	err := client.
		Get("http://httpbin.org/cookies",
			sreq.WithCookies(
				&http.Cookie{
					Name:  "n2",
					Value: "v2",
				},
			),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Cookies["n1"] != "v1" || resp.Cookies["n2"] != "v2" {
		t.Error("Client_AppendCookies test failed")
	}
}

func TestClient_SetBasicAuth(t *testing.T) {
	type response struct {
		Authenticated bool   `json:"authenticated"`
		User          string `json:"user"`
	}

	client := sreq.New().SetBasicAuth("admin", "pass")
	resp := new(response)
	err := client.
		Get("http://httpbin.org/basic-auth/admin/pass").
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Authenticated || resp.User != "admin" {
		t.Error("Client_SetBasicAuth test failed")
	}

	_resp := new(response)
	err = client.
		Get("http://httpbin.org/basic-auth/hello/world",
			sreq.WithBasicAuth("hello", "world"),
		).
		EnsureStatusOk().
		JSON(_resp)
	if err != nil {
		t.Fatal(err)
	}

	if !_resp.Authenticated || _resp.User != "hello" {
		t.Error("Client_SetBasicAuth test failed")
	}
}

func TestClient_SetBearerToken(t *testing.T) {
	type response struct {
		Authenticated bool   `json:"authenticated"`
		Token         string `json:"token"`
	}

	client := sreq.New().SetBearerToken("sreq")
	resp := new(response)
	err := client.
		Get("http://httpbin.org/bearer").
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Authenticated || resp.Token != "sreq" {
		t.Error("Client_SetBearerToken test failed")
	}

	_resp := new(response)
	err = client.
		Get("http://httpbin.org/bearer",
			sreq.WithBearerToken("hello"),
		).
		EnsureStatusOk().
		JSON(_resp)
	if err != nil {
		t.Fatal(err)
	}

	if !_resp.Authenticated || _resp.Token != "hello" {
		t.Error("Client_SetBearerToken test failed")
	}
}

func TestClient_SetContext(t *testing.T) {
	client := sreq.New().SetContext(nil)
	err := client.
		Get("http://httpbin.org/delay/10").
		Verbose(ioutil.Discard)
	if err == nil {
		t.Error("nil Context not checked")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	client = sreq.New().SetContext(ctx)
	err = client.
		Get("http://httpbin.org/delay/5",
			sreq.WithContext(context.Background()),
		).
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error("Client_SetContext test failed")
	}
}

func TestClient_SetRetry(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 3 {
			http.SetCookie(w, &http.Cookie{
				Name:  "uid",
				Value: "10086",
			})
		}
	}))
	defer ts.Close()

	condition := func(resp *sreq.Response) bool {
		_, err := resp.Cookie("uid")
		return err != nil
	}
	client := sreq.New().SetRetry(5, 1*time.Second, condition)
	cookie, err := client.
		Get(ts.URL).
		EnsureStatusOk().
		Cookie("uid")
	if err != nil {
		t.Fatal(err)
	}

	if cookie.Value != "10086" {
		t.Error("Client_SetRetry test failed")
	}
}

func TestClient_UseRequestInterceptors(t *testing.T) {
	logInterceptor := func(req *sreq.Request) error {
		rawRequest := req.RawRequest
		var w io.Writer = ioutil.Discard
		fmt.Fprintf(w, "> %s %s %s\r\n", rawRequest.Method, rawRequest.URL.RequestURI(), rawRequest.Proto)
		fmt.Fprintf(w, "> Host: %s\r\n", rawRequest.URL.Host)
		for k := range rawRequest.Header {
			fmt.Fprintf(w, "> %s: %s\r\n", k, rawRequest.Header.Get(k))
		}
		fmt.Fprint(w, ">\r\n")

		if rawRequest.GetBody != nil && rawRequest.ContentLength != 0 {
			rc, err := rawRequest.GetBody()
			if err != nil {
				return err
			}
			defer rc.Close()

			_, err = io.Copy(w, rc)
			if err != nil {
				return err
			}

			fmt.Fprint(w, "\r\n")
		}

		return nil
	}

	client := sreq.New().UseRequestInterceptors(logInterceptor)
	client.Post("http://httpbin.org/post",
		sreq.WithForm(sreq.Form{
			"k1": "v1",
		}),
		sreq.WithReferer("https://www.google.com"),
	)

	errMethodNotAllowed := errors.New("method not allowed")
	abortInterceptor := func(req *sreq.Request) error {
		rawRequest := req.RawRequest
		if rawRequest.Method == "DELETE" {
			return errMethodNotAllowed
		}

		return nil
	}

	client = sreq.New().UseRequestInterceptors(abortInterceptor)
	_, err := client.Delete("http://httpbin.org/delete",
		sreq.WithForm(sreq.Form{
			"uid": "10086",
		}),
	).Raw()
	if err != errMethodNotAllowed {
		t.Error("Client_UseRequestInterceptors test failed")
	}
}

func TestClient_UseResponseInterceptors(t *testing.T) {
	logInterceptor := func(resp *sreq.Response) error {
		var w io.Writer = ioutil.Discard
		rawResponse := resp.RawResponse
		fmt.Fprintf(w, "< %s %s\r\n", rawResponse.Proto, rawResponse.Status)
		for k := range rawResponse.Header {
			fmt.Fprintf(w, "< %s: %s\r\n", k, rawResponse.Header.Get(k))
		}
		fmt.Fprint(w, "<\r\n")

		fmt.Fprint(w, "< Cookies:\r\n")
		cookies, _ := resp.Cookies()
		for _, c := range cookies {
			fmt.Fprintf(w, "< %s: %s\r\n", c.Name, c.Value)
		}
		return nil
	}

	client := sreq.New().UseResponseInterceptors(logInterceptor)
	client.Get("https://www.baidu.com")

	errUnauthorized := errors.New("illegal user")
	abortInterceptor := func(resp *sreq.Response) error {
		rawResponse := resp.RawResponse
		if rawResponse.StatusCode == http.StatusUnauthorized {
			return errUnauthorized
		}

		return nil
	}

	client = sreq.New().UseResponseInterceptors(abortInterceptor)
	_, err := client.Get("http://httpbin.org/basic-auth/admin/pass",
		sreq.WithBasicAuth("user", "pass"),
	).Raw()
	if err != errUnauthorized {
		t.Error("Client_UseResponseInterceptors test failed")
	}
}

func TestClient_FilterCookie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "uid",
			Value: "10086",
		})
	}))
	defer ts.Close()

	client := sreq.New().DisableSession()
	cookie, err := client.FilterCookie(ts.URL, "uid")
	if err != sreq.ErrNilCookieJar {
		t.Error("Client_FilterCookie test failed")
	}

	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	client = sreq.New().SetCookieJar(jar)

	_, err = client.FilterCookies(ts.URL)
	if err != sreq.ErrJarCookiesNotPresent {
		t.Error("Client_FilterCookies test failed")
	}

	_, err = client.
		Get(ts.URL).
		EnsureStatusOk().
		Raw()
	if err != nil {
		t.Fatal(err)
	}

	cookie, err = client.FilterCookie(ts.URL, "uid")
	if err != nil {
		t.Fatal(err)
	}
	if got := cookie.Value; got != "10086" {
		t.Errorf("the cookie value expected to be: %s, but got: %s", "10086", got)
	}

	_, err = client.FilterCookie(ts.URL, "uuid")
	if err != sreq.ErrJarNamedCookieNotPresent {
		t.Error("Client_FilterCookie test failed")
	}
}

func TestClient_Do(t *testing.T) {
	req := sreq.NewRequest("GET", "http://httpbin.org/get")

	client := sreq.New()
	err := client.
		Do(req).
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func TestDefaultClient(t *testing.T) {
	rawClient, err := sreq.DefaultClient.Raw()
	if err != nil {
		t.Fatal(err)
	}

	transport, ok := rawClient.Transport.(*http.Transport)
	if !ok || transport == nil {
		t.Fatal("current transport isn't a non-nil *http.Transport instance")
	}

	req := sreq.NewRequest("GET", "http://httpbin.org/get")
	_, err = req.Raw()
	if err != nil {
		t.Fatal(err)
	}

	testDefaultClientDisableSession(t, rawClient)
	testDefaultClientSetCookieJar(t, rawClient)
	testDefaultClientGet(t)
	testDefaultClientHead(t)
	testDefaultClientPost(t)
	testDefaultClientPut(t)
	testDefaultClientPatch(t)
	testDefaultClientDelete(t)
	testDefaultClientSend(t)
	testDefaultClientDo(t, req)
	testDefaultClientFilterCookie(t)

	testDefaultClientDisableRedirect(t, rawClient)
	testDefaultClientSetRedirect(t, rawClient)
	testDefaultClientSetTimeout(t, rawClient)
	testDefaultClientSetProxyFromURL(t, req.RawRequest, transport)
	testDefaultClientDisableProxy(t, transport)
	testDefaultClientSetProxy(t, transport)
	testDefaultClientSetTLSClientConfig(t, transport)
	testDefaultClientAppendClientCertificates(t, transport)
	testDefaultClientAppendRootCAs(t, transport)
	testDefaultClientDisableVerify(t, transport)
	testDefaultClientSetTransport(t, rawClient)
}

func testDefaultClientGet(t *testing.T) {
	err := sreq.
		Get("http://httpbin.org/get").
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientHead(t *testing.T) {
	err := sreq.
		Head("http://httpbin.org").
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientPost(t *testing.T) {
	err := sreq.
		Post("http://httpbin.org/post").
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientPut(t *testing.T) {
	err := sreq.
		Put("http://httpbin.org/put").
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientPatch(t *testing.T) {
	err := sreq.
		Patch("http://httpbin.org/patch").
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientDelete(t *testing.T) {
	err := sreq.
		Delete("http://httpbin.org/delete").
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientSend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case sreq.MethodConnect, sreq.MethodOptions, sreq.MethodTrace:
			w.WriteHeader(http.StatusMethodNotAllowed)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	err := sreq.
		Send(sreq.MethodConnect, ts.URL).
		EnsureStatus(http.StatusMethodNotAllowed).
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}

	err = sreq.
		Send(sreq.MethodOptions, ts.URL).
		EnsureStatus(http.StatusMethodNotAllowed).
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}

	err = sreq.
		Send(sreq.MethodTrace, ts.URL).
		EnsureStatus(http.StatusMethodNotAllowed).
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientDo(t *testing.T, req *sreq.Request) {
	err := sreq.
		Do(req).
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientFilterCookie(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 3 {
			http.SetCookie(w, &http.Cookie{
				Name:  "uid",
				Value: "10086",
			})
		}
	}))
	defer ts.Close()

	sreq.SetHost("google.com")
	sreq.SetHeaders(sreq.Headers{})
	sreq.SetUserAgent("Go-http-client")
	sreq.SetReferer("https://www.google.com")
	sreq.SetCookies(&http.Cookie{
		Name:  "uid",
		Value: "10086",
	})
	sreq.SetBasicAuth("user", "pass")
	sreq.SetBearerToken("sreq")
	sreq.SetContext(context.Background())
	condition := func(resp *sreq.Response) bool {
		_, err := resp.Cookie("uid")
		return err != nil
	}
	sreq.SetRetry(5, 1*time.Second, condition)
	sreq.UseRequestInterceptors(func(req *sreq.Request) error {
		return nil
	})
	sreq.UseResponseInterceptors(func(resp *sreq.Response) error {
		return nil
	})

	err := sreq.
		Get(ts.URL).
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = sreq.FilterCookies(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = sreq.FilterCookie(ts.URL, "uid")
	if err != nil {
		t.Error(err)
	}
}

func testDefaultClientDisableSession(t *testing.T, rawClient *http.Client) {
	sreq.DisableSession()
	if rawClient.Jar != nil {
		t.Error("DefaultClient_DisableSession test failed")
	}
}

func testDefaultClientSetCookieJar(t *testing.T, rawClient *http.Client) {
	jar, _ := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})

	sreq.SetCookieJar(jar)
	if rawClient.Jar != jar {
		t.Error("DefaultClient_SetCookieJar test failed")
	}
}

func testDefaultClientDisableRedirect(t *testing.T, rawClient *http.Client) {
	sreq.DisableRedirect()
	if err := rawClient.CheckRedirect(nil, nil); err != http.ErrUseLastResponse {
		t.Error("DefaultClient_DisableRedirect test failed")
	}
}

func testDefaultClientSetRedirect(t *testing.T, rawClient *http.Client) {
	sreq.SetRedirect(nil)
	if rawClient.CheckRedirect != nil {
		t.Error("DefaultClient_SetRedirect test failed")
	}
}

func testDefaultClientSetTimeout(t *testing.T, rawClient *http.Client) {
	const (
		timeout = 3 * time.Second
	)

	sreq.SetTimeout(timeout)
	if got := rawClient.Timeout; got != timeout {
		t.Error("DefaultClient_SetTimeout test failed")
	}
}

func testDefaultClientSetProxyFromURL(t *testing.T, rawRequest *http.Request, transport *http.Transport) {
	const (
		url = "socks5://127.0.0.1:1080"
	)

	sreq.SetProxyFromURL(url)
	proxyURL, err := transport.Proxy(rawRequest)
	if err != nil {
		t.Fatal(err)
	}
	if got := proxyURL.String(); got != url {
		t.Error("DefaultClient_SetProxyFromURL test failed")
	}
}

func testDefaultClientDisableProxy(t *testing.T, transport *http.Transport) {
	sreq.DisableProxy()
	if transport.Proxy != nil {
		t.Error("DefaultClient_DisableProxy test failed")
	}
}

func testDefaultClientSetProxy(t *testing.T, transport *http.Transport) {
	sreq.SetProxy(http.ProxyFromEnvironment)
	if transport.Proxy == nil {
		t.Error("DefaultClient_SetProxy test failed")
	}
}

func testDefaultClientSetTLSClientConfig(t *testing.T, transport *http.Transport) {
	sreq.SetTLSClientConfig(&tls.Config{})
	if transport.TLSClientConfig == nil {
		t.Fatal("DefaultClient_SetTLSClientConfig test failed")
	}
}

func testDefaultClientAppendClientCertificates(t *testing.T, transport *http.Transport) {
	sreq.AppendClientCertificates(tls.Certificate{})
	if transport.TLSClientConfig == nil || len(transport.TLSClientConfig.Certificates) != 1 {
		t.Error("DefaultClient_AppendClientCertificates test failed")
	}
}

func testDefaultClientAppendRootCAs(t *testing.T, transport *http.Transport) {
	const (
		pemFileExist = "./testdata/root-ca.pem"
	)

	sreq.AppendRootCAs(pemFileExist)
	if transport.TLSClientConfig == nil || transport.TLSClientConfig.RootCAs == nil {
		t.Error("DefaultClient_AppendRootCAs test failed")
	}
}

func testDefaultClientDisableVerify(t *testing.T, transport *http.Transport) {
	sreq.DisableVerify()
	if transport.TLSClientConfig == nil || !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("DefaultClient_DisableVerify test failed")
	}
}

func testDefaultClientSetTransport(t *testing.T, rawClient *http.Client) {
	sreq.SetTransport(nil)
	if rawClient.Transport != nil {
		t.Error("DefaultClient_SetTransport test failed")
	}
}
