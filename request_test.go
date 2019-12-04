package sreq_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/winterssy/sreq"
)

func TestNewRequest(t *testing.T) {
	const (
		invalidMethod = "@"
	)

	_, err := sreq.
		NewRequest(invalidMethod, "http://httpbin.org/post").
		SetBody(nil).
		SetHost("httpbin.org").
		SetHeaders(sreq.Headers{
			"Referer": "http://httpbin.org",
		}).
		SetUserAgent("Go-http-client").
		SetReferer("https://www.google.com").
		SetCookies(
			&http.Cookie{
				Name:  "n1",
				Value: "v1",
			},
			&http.Cookie{
				Name:  "n2",
				Value: "v2",
			},
		).
		SetQuery(sreq.Params{
			"k1": "v1",
		}).
		SetContent([]byte("hello world")).
		SetContentType("text/plain").
		SetText("hello world").
		SetForm(sreq.Form{
			"k2": "v2",
		}).
		SetJSON(sreq.JSON{
			"msg": "hi&hello",
			"num": 2019,
		}, true).
		SetFiles(sreq.Files{
			"field1": "./testdata/testfile1.txt",
			"field2": "./testdata/testfile2.txt",
		}).
		SetBasicAuth("user", "pass").
		SetBearerToken("sreq").
		SetContext(context.Background()).
		SetTimeout(3*time.Second).
		SetRetry(3, 1*time.Second, 2*time.Minute).
		Raw()
	if err == nil {
		t.Error("NewRequest test failed")
	}
}

func TestWithBody(t *testing.T) {
	body := bytes.NewBuffer([]byte{})
	client := sreq.New()
	err := client.
		Post("http://httpbin.org/post",
			sreq.WithBody(body),
			sreq.WithContentType("text/plain"),
		).Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func TestWithQuery(t *testing.T) {
	type response struct {
		Args map[string]string `json:"args"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get",
			sreq.WithQuery(sreq.Params{
				"k1": "v1",
				"k2": "v2",
			}),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Args["k1"] != "v1" || resp.Args["k2"] != "v2" {
		t.Error("WithQuery test failed")
	}
}

func TestWithHost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Host))
	}))
	defer ts.Close()

	client := sreq.New()
	data, err := client.
		Get(ts.URL,
			sreq.WithHost("github.com"),
		).
		EnsureStatusOk().
		Text()
	if err != nil {
		t.Fatal(err)
	}

	if data != "github.com" {
		t.Error("WithHost test failed")
	}
}

func TestWithHeaders(t *testing.T) {
	type response struct {
		Headers map[string]string `json:"headers"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get",
			sreq.WithHeaders(sreq.Headers{
				"Origin":  "http://httpbin.org",
				"Referer": "http://httpbin.org",
			}),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Headers["Origin"] != "http://httpbin.org" || resp.Headers["Referer"] != "http://httpbin.org" {
		t.Error("WithHeaders test failed")
	}
}

func TestWithUserAgent(t *testing.T) {
	type response struct {
		Headers map[string]string `json:"headers"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get",
			sreq.WithUserAgent("Go-http-client"),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Headers["User-Agent"] != "Go-http-client" {
		t.Error("WithUserAgent test failed")
	}
}

func TestWithReferer(t *testing.T) {
	type response struct {
		Headers map[string]string `json:"headers"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get",
			sreq.WithReferer("https://www.google.com"),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Headers["Referer"] != "https://www.google.com" {
		t.Error("WithReferer test failed")
	}
}

func TestWithCookies(t *testing.T) {
	type response struct {
		Cookies map[string]string `json:"cookies"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Get("http://httpbin.org/cookies",
			sreq.WithCookies(
				&http.Cookie{
					Name:  "n1",
					Value: "v1",
				},
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
		t.Error("WithCookies test failed")
	}
}

func TestWithContent(t *testing.T) {
	client := sreq.New()
	err := client.
		Post("http://httpbin.org/post",
			sreq.WithContent([]byte("hello world")),
			sreq.WithContentType("text/plain"),
		).
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}

	type response struct {
		Data string `json:"data"`
	}

	resp := new(response)
	err = client.
		Post("http://httpbin.org/post",
			sreq.WithContent([]byte("hello world")),
			sreq.WithContentType("text/plain"),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Data != "hello world" {
		t.Error("WithContent test failed")
	}
}

func TestWithText(t *testing.T) {
	client := sreq.New()
	err := client.
		Post("http://httpbin.org/post",
			sreq.WithText("hello world"),
		).
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}

	type response struct {
		Data string `json:"data"`
	}

	resp := new(response)
	err = client.
		Post("http://httpbin.org/post",
			sreq.WithText("hello world"),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Data != "hello world" {
		t.Error("WithText test failed")
	}
}

func TestWithForm(t *testing.T) {
	type response struct {
		Form map[string]string `json:"form"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Post("http://httpbin.org/post",
			sreq.WithForm(sreq.Form{
				"k1": "v1",
				"k2": "v2",
			}),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Form["k1"] != "v1" || resp.Form["k2"] != "v2" {
		t.Error("WithForm test failed")
	}
}

func TestWithJSON(t *testing.T) {
	client := sreq.New()
	err := client.
		Post("http://httpbin.org/post",
			sreq.WithJSON(sreq.JSON{
				"num": math.Inf(1),
			}, true),
		).
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err == nil {
		t.Error("WithJSON test failed")
	}

	err = client.
		Post("http://httpbin.org/post",
			sreq.WithJSON(sreq.JSON{
				"msg": "hi&hello",
				"num": 2019,
			}, true),
		).
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}

	type response struct {
		JSON struct {
			Msg string `json:"msg"`
			Num int    `json:"num"`
		} `json:"json"`
	}

	resp := new(response)
	err = client.
		Post("http://httpbin.org/post",
			sreq.WithJSON(sreq.JSON{
				"msg": "hi&hello",
				"num": 2019,
			}, true),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.JSON.Msg != "hi&hello" || resp.JSON.Num != 2019 {
		t.Error("WithJSON test failed")
	}

	_resp := new(response)
	err = client.
		Post("http://httpbin.org/post",
			sreq.WithJSON(sreq.JSON{
				"msg": "hi&hello",
				"num": 2019,
			}, false),
		).
		EnsureStatusOk().
		JSON(_resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.JSON.Msg != "hi&hello" || resp.JSON.Num != 2019 {
		t.Error("WithJSON test failed")
	}
}

func TestWithFiles(t *testing.T) {
	client := sreq.New()
	_, err := client.
		Post("http://httpbin.org/post",
			sreq.WithFiles(sreq.Files{
				"field": "./testdata/file_not_exist.txt",
			}),
		).
		EnsureStatusOk().
		Raw()
	if err == nil {
		t.Error("WithFiles test failed")
	}

	_, err = client.
		Post("http://httpbin.org/post",
			sreq.WithFiles(sreq.Files{
				"field": "./testdata",
			}),
		).
		EnsureStatusOk().
		Raw()
	if err == nil {
		t.Error("WithFiles test failed")
	}

	type response struct {
		Files map[string]string `json:"files"`
	}

	resp := new(response)
	err = client.
		Post("http://httpbin.org/post",
			sreq.WithFiles(sreq.Files{
				"field1": "./testdata/testfile1.txt",
				"field2": "./testdata/testfile2.txt",
			}),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Files["field1"] != "testfile1.txt" || resp.Files["field2"] != "testfile2.txt" {
		t.Error("WithFiles test failed")
	}
}

func TestWithBasicAuth(t *testing.T) {
	type response struct {
		Authenticated bool   `json:"authenticated"`
		User          string `json:"user"`
	}

	resp := new(response)
	client := sreq.New()
	err := client.
		Get("http://httpbin.org/basic-auth/admin/pass",
			sreq.WithBasicAuth("admin", "pass"),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Authenticated || resp.User != "admin" {
		t.Error("WithBasicAuth test failed")
	}
}

func TestWithBearerToken(t *testing.T) {
	type response struct {
		Authenticated bool   `json:"authenticated"`
		Token         string `json:"token"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Get("http://httpbin.org/bearer",
			sreq.WithBearerToken("sreq"),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Authenticated || resp.Token != "sreq" {
		t.Error("WithBearerToken test failed")
	}
}

func TestWithContext(t *testing.T) {
	client := sreq.New()
	err := client.
		Get("http://httpbin.org/delay/10",
			sreq.WithContext(nil),
		).
		Verbose(ioutil.Discard)
	if err == nil {
		t.Error("nil Context not checked")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = client.
		Get("http://httpbin.org/delay/10",
			sreq.WithContext(ctx),
		).
		Verbose(ioutil.Discard)
	if err == nil {
		t.Error("WithContext test failed")
	}
}

func TestWithTimeout(t *testing.T) {
	client := sreq.New()
	err := client.
		Get("http://httpbin.org/delay/10",
			sreq.WithTimeout(3*time.Second),
		).
		Verbose(ioutil.Discard)
	if err == nil {
		t.Error("WithTimeout test failed")
	}
}

func TestWithRetry(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 5 {
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

	client := sreq.New().SetRetry(3, 1*time.Second, 2*time.Minute, condition)
	cookie, err := client.
		Get(ts.URL,
			sreq.WithRetry(5, 1*time.Second, 3*time.Second, condition),
		).
		EnsureStatusOk().
		Cookie("uid")
	if err != sreq.ErrRetryMaxDurationExceeded {
		t.Error("WithRetry test failed")
	}

	cookie, err = client.
		Get(ts.URL,
			sreq.WithRetry(5, 1*time.Second, 2*time.Minute, condition),
		).
		EnsureStatusOk().
		Cookie("uid")
	if err != nil {
		t.Fatal(err)
	}

	if cookie.Value != "10086" {
		t.Error("WithRetry test failed")
	}

	attempts = 0
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.
		Get(ts.URL,
			sreq.WithContext(ctx),
			sreq.WithRetry(5, 1*time.Second, 2*time.Minute, condition),
		).
		EnsureStatusOk().
		Cookie("uid")
	if err == nil {
		t.Error("context should have priority over the retry policy")
	}
}
