package sreq_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/winterssy/sreq"
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
		switch r.Method {
		case sreq.MethodPost:
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "created")
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, "method not allowed")
		}
	}))
	defer ts.Close()

	client := sreq.New()
	data, err := client.
		Post(ts.URL).
		EnsureStatusOk().
		Text()
	if err != nil || data != "created" {
		t.Error(err)
	}

	data, err = client.
		Put(ts.URL).
		EnsureStatus2xx().
		Text()
	if err == nil || data != "" {
		t.Error("Response_Text test failed")
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

	_, err = r.Text()
	if err != nil {
		t.Error(err)
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
