package sreq_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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
		SetJSON(map[string]interface{}{
			"msg": "hi&hello",
			"num": 2019,
		}, true).
		SetXML(nil).
		SetMultipart(nil, nil).
		SetBasicAuth("user", "pass").
		SetBearerToken("sreq").
		SetContext(context.Background()).
		SetTimeout(3*time.Second).
		SetRetry(3, 1*time.Second).
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
		Args testStruct `json:"args"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Get("http://httpbin.org/get",
			sreq.WithQuery(testValues),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Args.Int != "2019" || resp.Args.String != "2019" ||
		!reflect.DeepEqual(resp.Args.StringArray, testStringArray) ||
		!reflect.DeepEqual(resp.Args.IntArray, testStringArray) ||
		!reflect.DeepEqual(resp.Args.StringIntArray, testStringArray) {
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
				"string":           "2019",
				"int":              2019,
				"string-array":     testStringArray,
				"int-array":        []int{10086, 10010, 10000},
				"string-int-array": []interface{}{"10086", 10010, 10000},
			}),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	want := "10086,10010,10000"
	if resp.Headers["String"] != "2019" || resp.Headers["Int"] != "2019" ||
		resp.Headers["String-Array"] != want ||
		resp.Headers["Int-Array"] != want ||
		resp.Headers["String-Int-Array"] != want {
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
		Form testStruct `json:"form"`
	}

	client := sreq.New()
	resp := new(response)
	err := client.
		Post("http://httpbin.org/post",
			sreq.WithForm(testValues),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Form.Int != "2019" || resp.Form.String != "2019" ||
		!reflect.DeepEqual(resp.Form.StringArray, testStringArray) ||
		!reflect.DeepEqual(resp.Form.IntArray, testStringArray) ||
		!reflect.DeepEqual(resp.Form.StringIntArray, testStringArray) {
		t.Error("WithForm test failed")
	}
}

func TestWithJSON(t *testing.T) {
	client := sreq.New()
	err := client.
		Post("http://httpbin.org/post",
			sreq.WithJSON(map[string]interface{}{
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
			sreq.WithJSON(map[string]interface{}{
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
			sreq.WithJSON(map[string]interface{}{
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
			sreq.WithJSON(map[string]interface{}{
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

func TestWithXML(t *testing.T) {
	type plant struct {
		XMLName xml.Name `xml:"plant"`
		Id      int      `xml:"id,attr"`
		Name    string   `xml:"name"`
		Origin  []string `xml:"origin"`
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data plant
		err := xml.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(data)
	}))
	defer ts.Close()

	client := sreq.New()
	_, err := client.
		Post(ts.URL,
			sreq.WithXML(make(map[string]interface{})),
		).
		EnsureStatusOk().
		Raw()
	if _, ok := err.(*sreq.RequestError); !ok {
		t.Error("WithXML test failed")
	}

	origin := []string{"Ethiopia", "Brazil"}
	coffee := &plant{
		Id:     27,
		Name:   "Coffee",
		Origin: origin,
	}

	resp := client.
		Post(ts.URL,
			sreq.WithXML(coffee),
		).
		EnsureStatusOk()

	result := new(plant)
	err = resp.XML(result)
	if err != nil {
		t.Fatal(err)
	}
	if result.Id != 27 || result.Name != "Coffee" || !reflect.DeepEqual(result.Origin, origin) {
		t.Error("Response_XML test failed")
	}

	_result := new(plant)
	err = resp.XML(_result)
	if err != nil {
		t.Fatal(err)
	}
	if _result.Id != 27 || _result.Name != "Coffee" || !reflect.DeepEqual(_result.Origin, origin) {
		t.Error("Response_ReuseBody test failed")
	}
}

type errBody struct{}

func (e *errBody) Read(p []byte) (int, error) {
	return 0, errors.New("no body")
}

func TestWithMultipart(t *testing.T) {
	// For Charles
	// client := sreq.New().SetProxyFromURL("http://127.0.0.1:7777")

	client := sreq.New()
	_, err := client.
		Post("http://httpbin.org/post",
			sreq.WithMultipart(sreq.Files{
				"file": sreq.NewFile("errorBody", &errBody{}),
			}, nil)).
		Raw()
	if _, ok := err.(*sreq.RequestError); !ok {
		t.Error("WithMultipart test failed")
	}

	_, err = client.
		Post("http://httpbin.org/post",
			sreq.WithMultipart(sreq.Files{
				"file": &sreq.File{
					Body: strings.NewReader("Filename not specified, sreq will raise an error and abort request"),
				},
			}, nil)).
		Raw()
	if _, ok := err.(*sreq.RequestError); !ok {
		t.Error("WithMultipart test failed")
	}

	files := sreq.Files{
		"file1": sreq.
			MustOpen("./testdata/testfile1.txt"),
		"file2": sreq.
			MustOpen("./testdata/testfile2.txt").
			SetFilename("testfile2.txt"),
		"file3": sreq.
			NewFile("testfile3.txt",
				bytes.NewReader([]byte("<p>This is a text file from memory</p>"))).
			SetMIME("text/html; charset=utf-8"),
	}

	type response struct {
		Files map[string]string `json:"files"`
		Form  testStruct        `json:"form"`
	}

	resp := new(response)
	err = client.
		Post("http://httpbin.org/post",
			sreq.WithMultipart(files, testValues),
		).
		EnsureStatusOk().
		JSON(resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Files["file1"] != "testfile1.txt" || resp.Files["file2"] != "testfile2.txt" ||
		resp.Files["file3"] != "<p>This is a text file from memory</p>" ||
		resp.Form.Int != "2019" || resp.Form.String != "2019" ||
		!reflect.DeepEqual(resp.Form.StringArray, testStringArray) ||
		!reflect.DeepEqual(resp.Form.IntArray, testStringArray) ||
		!reflect.DeepEqual(resp.Form.StringIntArray, testStringArray) {
		t.Error("WithMultipart test failed")
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

	client := sreq.New().SetRetry(3, 1*time.Second, condition)
	cookie, err := client.
		Get(ts.URL,
			sreq.WithRetry(5, 1*time.Second, condition),
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
			sreq.WithRetry(5, 1*time.Second, condition),
		).
		EnsureStatusOk().
		Cookie("uid")
	if err == nil {
		t.Error("context should have priority over the retry policy")
	}
}
