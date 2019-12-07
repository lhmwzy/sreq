package sreq_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/winterssy/sreq"
)

func TestParams(t *testing.T) {
	p := make(sreq.Params)

	p.Set("k1", "v1")
	p.Set("k2", "v2")
	p.Set("k3", "v3")
	if p["k1"] != "v1" || p["k2"] != "v2" || p["k3"] != "v3" {
		t.Fatal("Params_Set test failed")
	}

	if p.Get("k1") != "v1" || p.Get("k2") != "v2" || p.Get("k3") != "v3" {
		t.Error("Params_Get test failed")
	}

	p.Del("k1")
	if p["k1"] != nil || len(p) != 2 {
		t.Error("Params_Del test failed")
	}

	want := "k2=v2&k3=v3"
	if got := p.String(); got != want {
		t.Errorf("Params_String got: %s, want: %s", got, want)
	}

	p = sreq.Params{
		"e": "user/pass",
	}
	want = "e=user/pass"
	if got := p.Encode(); got != want {
		t.Errorf("Params_Encode got: %s, want: %s", got, want)
	}

	p = sreq.Params{
		"string":      "2019",
		"int":         2019,
		"stringArray": []string{"10086", "10010", "10000"},
		"intArray":    []int{10086, 10010, 10000},
	}
	want = "int=2019&intArray=10086&intArray=10010&intArray=10000&" +
		"string=2019&stringArray=10086&stringArray=10010&stringArray=10000"
	if got := p.Encode(); got != want {
		t.Errorf("Params_Encode got: %s, want: %s", got, want)
	}
}

func TestHeaders(t *testing.T) {
	h1 := make(sreq.Headers)

	h1.Set("k1", "v1")
	h1.Set("k2", "v2")
	if h1["k1"] != "v1" || h1["k2"] != "v2" {
		t.Fatal("Headers_Set test failed")
	}

	if h1.Get("k1") != "v1" || h1.Get("k2") != "v2" {
		t.Error("Headers_Get test failed")
	}

	h1.Del("k1")
	if h1["k1"] != nil || len(h1) != 1 {
		t.Error("Headers_Del test failed")
	}

	h2 := make(sreq.Headers)
	if err := json.Unmarshal([]byte(h1.String()), &h2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(h2, h1) {
		t.Error("Headers_String test failed")
	}
}

func TestForm(t *testing.T) {
	f := make(sreq.Form)

	f.Set("k1", "v1")
	f.Set("k2", "v2")
	f.Set("k3", "v3")
	if f["k1"] != "v1" || f["k2"] != "v2" || f["k3"] != "v3" {
		t.Fatal("Form_Set test failed")
	}

	if f.Get("k1") != "v1" || f.Get("k2") != "v2" || f.Get("k3") != "v3" {
		t.Error("Form_Get test failed")
	}

	f.Del("k1")
	if f["k1"] != nil || len(f) != 2 {
		t.Error("Form_Del test failed")
	}

	want := "k2=v2&k3=v3"
	if got := f.String(); got != want {
		t.Errorf("Form_String got: %s, want: %s", got, want)
	}

	f = sreq.Form{
		"q":       []string{"Go语言", "Python"},
		"offset":  0,
		"limit":   100,
		"invalid": []interface{}{"hello", 2019},
	}
	want = "limit=100&offset=0&q=Go语言&q=Python"
	if got := f.Encode(); got != want {
		t.Errorf("Form_Encode got: %s, want: %s", got, want)
	}
}

func TestJSON(t *testing.T) {
	j := make(sreq.JSON)

	j.Set("msg", "hello world")
	j.Set("num", 2019)
	if j["msg"] != "hello world" || j["num"] != 2019 {
		t.Fatal("JSON_Set test failed")
	}

	if j.Get("msg") != "hello world" || j.Get("num") != 2019 {
		t.Error("JSON_Get test failed")
	}

	j.Del("msg")
	if j["msg"] != nil || len(j) != 1 {
		t.Error("JSON_Del test failed")
	}

	want := "{\n\t\"num\": 2019\n}\n"
	if got := j.String(); got != want {
		t.Errorf("JSON_string got: %q, want: %q", got, want)
	}
}

func TestFiles(t *testing.T) {
	f := make(sreq.Files)

	f.Set("file1", &sreq.FileForm{
		Reader: &os.File{},
		MIME:   "image/png",
	})
	f.Set("file2", &sreq.FileForm{
		Reader: strings.NewReader("hello world"),
		MIME:   "text/plain",
	})

	if len(f) != 2 {
		t.Error("Files_Set test failed")
	}

	file1 := f.Get("file1")
	if file1.MIME != "image/png" {
		t.Error("Files_Get test failed")
	}

	f.Del("file2")
	if f.Get("file2") != nil || len(f) != 1 {
		t.Error("Files_Del test failed")
	}
}
