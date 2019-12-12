package sreq_test

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/winterssy/sreq"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	testFileName = "testdata.json"
)

func TestResponse_Resolve(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	client := sreq.New()
	resp, err := client.
		Send(sreq.MethodGet, ts.URL).
		Raw()
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Response_Resolve got: %d, want: %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestResponse_Text(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var _w io.Writer
		q := r.URL.Query().Get("e")
		switch {
		case strings.EqualFold(q, "UTF-8"):
			_w = transform.NewWriter(w, unicode.UTF8.NewEncoder())
		case strings.EqualFold(q, "GBK"):
			_w = transform.NewWriter(w, simplifiedchinese.GBK.NewEncoder())
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_w.Write([]byte("你好世界"))
	}))
	defer ts.Close()

	client := sreq.New()
	want := "你好世界"
	data, err := client.
		Get(ts.URL,
			sreq.WithQuery(sreq.Params{
				"e": "utf-8",
			}),
		).
		EnsureStatusOk().
		Text()
	if err != nil || data != want {
		t.Errorf("Response_Text got: %q, want: %q", data, want)
	}

	data, err = client.
		Get(ts.URL,
			sreq.WithQuery(sreq.Params{
				"e": "gbk",
			}),
		).
		EnsureStatus2xx().
		Text(simplifiedchinese.GBK)
	if err != nil || data != want {
		t.Errorf("Response_Text got: %q, want: %q", data, want)
	}
}

func TestResponse_JSON(t *testing.T) {
	data := make(map[string]interface{})
	client := sreq.New()
	err := client.
		Get("https://www.google.com/404").
		EnsureStatusOk().
		JSON(&data)
	if err == nil {
		t.Error("Response_JSON test failed")
	}
}

func TestResponse_H(t *testing.T) {
	client := sreq.New()
	h, err := client.
		Get("http://httpbin.org/get",
			sreq.WithQuery(sreq.Params{
				"k1": "v1",
				"k2": "v2",
			}),
		).
		EnsureStatusOk().
		H()
	if err != nil {
		t.Fatal(err)
	}

	args := h.GetH("args")
	if args.GetString("k1") != "v1" || args.GetString("k2") != "v2" {
		t.Error("Response_H test failed")
	}

	h, err = client.
		Post("http://httpbin.org/post",
			sreq.WithJSON(map[string]interface{}{
				"songs": []map[string]interface{}{
					{
						"id":     29947420,
						"name":   "Fade",
						"artist": "Alan Walker",
						"album":  "Fade",
					},
					{
						"id":     444269135,
						"name":   "Alone",
						"artist": "Alan Walker",
						"album":  "Alone",
					},
				},
			}, true),
		).
		EnsureStatusOk().
		H()
	if err != nil {
		t.Fatal(err)
	}

	data := h.GetH("json").GetHSlice("songs")
	if len(data) != 2 {
		t.Fatal("Response_H test failed")
	}

	fade := data[0]
	if fade.GetInt("id") != 29947420 || fade.GetString("name") != "Fade" ||
		fade.GetString("artist") != "Alan Walker" {
		t.Error("Response_H test failed")
	}
}

func TestResponse_XML(t *testing.T) {
	type plant struct {
		XMLName xml.Name `xml:"plant"`
		Id      int      `xml:"id,attr"`
		Name    string   `xml:"name"`
		Origin  []string `xml:"origin"`
	}

	var data plant
	client := sreq.New()
	err := client.
		Get("https://www.google.com/404").
		EnsureStatusOk().
		XML(&data)
	if err == nil {
		t.Error("Response_XML test failed")
	}
}

func TestResponse_Cookie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "uid",
			Value: "10086",
		})
	}))
	defer ts.Close()

	client := sreq.New()
	resp := client.
		Get(ts.URL).
		EnsureStatusOk()

	cookie, err := resp.Cookie("uid")
	if err != nil {
		t.Fatal(err)
	}
	if cookie.Value != "10086" {
		t.Error("Response_Cookie test failed")
	}

	_, err = resp.Cookie("uuid")
	if err != sreq.ErrResponseNamedCookieNotPresent {
		t.Error("Response_Cookie test failed")
	}
}

func TestResponse_EnsureStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case sreq.MethodGet:
			w.WriteHeader(http.StatusOK)
		case sreq.MethodPost:
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer ts.Close()

	client := sreq.New()
	_, err := client.
		Get(ts.URL).
		EnsureStatusOk().
		Raw()
	if err != nil {
		t.Error(err)
	}

	_, err = client.
		Post(ts.URL).
		EnsureStatus2xx().
		Raw()
	if err != nil {
		t.Error(err)
	}

	_, err = client.
		Patch(ts.URL).
		EnsureStatus2xx().
		Raw()
	if err == nil {
		t.Error("Response_EnsureStatus2xx test failed")
	}

	_, err = client.
		Patch(ts.URL).
		EnsureStatusOk().
		EnsureStatus2xx().
		Raw()
	if err == nil {
		t.Error("Response_EnsureStatus2xx test failed")
	}

	_, err = client.
		Delete(ts.URL).
		EnsureStatus(http.StatusForbidden).
		Raw()
	if err != nil {
		t.Error(err)
	}

	_, err = client.
		Delete(ts.URL).
		EnsureStatus(http.StatusOK).
		Raw()
	if err == nil {
		t.Error("Response_EnsureStatus test failed")
	}
}

func TestResponse_Save(t *testing.T) {
	client := sreq.New()
	err := client.
		Get("https://www.google.com/404").
		EnsureStatusOk().
		Save(testFileName, 0664)
	if err == nil {
		t.Error("Response_Save test failed")
	}

	err = client.
		Get("http://httpbin.org/get").
		EnsureStatusOk().
		Save(testFileName, 0664)
	if err != nil {
		t.Error(err)
	}
}

func TestResponse_Verbose(t *testing.T) {
	client := sreq.New()
	err := client.
		Get("http://httpbin.org/get",
			sreq.WithQuery(sreq.Params{
				"uid": "10086",
			}),
		).
		EnsureStatusOk().
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}

	err = client.
		Post("http://httpbin.org/post",
			sreq.WithForm(sreq.Form{
				"uid": "10086",
			}),
		).
		Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}

func TestResponse_ReuseBody(t *testing.T) {
	type response struct {
		Args map[string]string `json:"args"`
	}

	client := sreq.New()
	r := client.
		Get("http://httpbin.org/get",
			sreq.WithQuery(sreq.Params{
				"k1": "v1",
				"k2": "v2",
			}),
		)

	_, err := r.Content()
	if err != nil {
		t.Fatal(err)
	}

	data, err := r.Text()
	if err != nil {
		t.Error(err)
	}
	if data == "" {
		t.Error("Response_ReuseBody test failed")
	}

	resp := new(response)
	err = r.JSON(resp)
	if err != nil {
		t.Error(err)
	}
	if resp.Args["k1"] != "v1" || resp.Args["k2"] != "v2" {
		t.Error("Response_ReuseBody test failed")
	}

	err = r.Save(testFileName, 0664)
	if err != nil {
		t.Error(err)
	}

	err = r.Verbose(ioutil.Discard)
	if err != nil {
		t.Error(err)
	}
}
