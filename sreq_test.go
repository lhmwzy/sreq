package sreq_test

import (
	"encoding/json"
	"reflect"
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
	if p["k1"] != "" || len(p) != 2 {
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
}

func TestParams_Clone(t *testing.T) {
	var p1 sreq.Params
	p2 := p1.Clone()
	if p2 != nil {
		t.Error("Params_Clone test failed")
	}

	p1 = sreq.Params{
		"k": "v",
	}
	p2 = p1.Clone()
	if &p2 == &p1 || !reflect.DeepEqual(p2, p1) {
		t.Error("Params_Clone test failed")
	}
}

func TestParams_Merge(t *testing.T) {
	var p1 sreq.Params
	p2 := sreq.Params{
		"k": "v",
	}
	p3 := p1.Merge(p2)
	if &p3 == &p1 || &p3 == &p2 || !reflect.DeepEqual(p3, p2) {
		t.Error("Params_Merge test failed")
	}

	p1 = sreq.Params{
		"k1": "v1",
	}
	p2 = sreq.Params{
		"k2": "v2",
	}
	p3 = p1.Merge(p2)
	if &p3 == &p1 || &p3 == &p2 || len(p3) != 2 || p3["k2"] != "v2" {
		t.Error("Params_Merge test failed")
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
	if h1["k1"] != "" || len(h1) != 1 {
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

func TestHeaders_Clone(t *testing.T) {
	var h1 sreq.Headers
	h2 := h1.Clone()
	if h2 != nil {
		t.Error("Headers_Clone test failed")
	}

	h1 = sreq.Headers{
		"k": "v",
	}
	h2 = h1.Clone()
	if &h2 == &h1 || !reflect.DeepEqual(h2, h1) {
		t.Error("Headers_Clone test failed")
	}
}

func TestHeaders_Merge(t *testing.T) {
	var h1 sreq.Headers
	h2 := sreq.Headers{
		"k": "v",
	}
	h3 := h1.Merge(h2)
	if &h3 == &h1 || &h3 == &h2 || !reflect.DeepEqual(h3, h2) {
		t.Error("Headers_Merge test failed")
	}

	h1 = sreq.Headers{
		"k1": "v1",
	}
	h2 = sreq.Headers{
		"k2": "v2",
	}
	h3 = h1.Merge(h2)
	if &h3 == &h1 || &h3 == &h2 || len(h3) != 2 || h3["k2"] != "v2" {
		t.Error("Headers_Merge test failed")
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
	if f["k1"] != "" || len(f) != 2 {
		t.Error("Form_Del test failed")
	}

	want := "k2=v2&k3=v3"
	if got := f.String(); got != want {
		t.Errorf("Form_String got: %s, want: %s", got, want)
	}

	f = sreq.Form{
		"q":      "Go语言",
		"offset": "0",
		"limit":  "100",
	}
	want = "limit=100&offset=0&q=Go语言"
	if got := f.Encode(); got != want {
		t.Errorf("Form_Encode got: %s, want: %s", got, want)
	}
}

func TestForm_Clone(t *testing.T) {
	var f1 sreq.Form
	f2 := f1.Clone()
	if f2 != nil {
		t.Error("Form_Clone test failed")
	}

	f1 = sreq.Form{
		"k": "v",
	}
	f2 = f1.Clone()
	if &f2 == &f1 || !reflect.DeepEqual(f2, f1) {
		t.Error("Form_Clone test failed")
	}
}

func TestForm_Merge(t *testing.T) {
	var f1 sreq.Form
	f2 := sreq.Form{
		"k": "v",
	}
	f3 := f1.Merge(f2)
	if &f3 == &f1 || &f3 == &f2 || !reflect.DeepEqual(f3, f2) {
		t.Error("Form_Merge test failed")
	}

	f1 = sreq.Form{
		"k1": "v1",
	}
	f2 = sreq.Form{
		"k2": "v2",
	}
	f3 = f1.Merge(f2)
	if &f3 == &f1 || &f3 == &f2 || len(f3) != 2 || f3["k2"] != "v2" {
		t.Error("Form_Merge test failed")
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

func TestJSON_Clone(t *testing.T) {
	var j1 sreq.JSON
	j2 := j1.Clone()
	if j2 != nil {
		t.Error("JSON_Clone test failed")
	}

	j1 = sreq.JSON{
		"k": "v",
	}
	j2 = j1.Clone()
	if &j2 == &j1 || len(j2) != 1 || j1["k"] != "v" {
		t.Error("JSON_Clone test failed")
	}
}

func TestJSON_Merge(t *testing.T) {
	var j1 sreq.JSON
	j2 := sreq.JSON{
		"offset": 1,
	}
	j3 := j1.Merge(j2)
	if &j3 == &j1 || &j3 == &j2 || len(j3) != 1 || j3["offset"] != 1 {
		t.Error("JSON_Merge test failed")
	}

	j1 = sreq.JSON{
		"msg": "hello world",
	}
	j2 = sreq.JSON{
		"num": 2019,
	}
	j3 = j1.Merge(j2)
	if &j3 == &j1 || &j3 == &j2 || len(j3) != 2 || j3["num"] != 2019 {
		t.Error("JSON_Merge test failed")
	}
}

func TestFiles(t *testing.T) {
	f1 := make(sreq.Files)

	f1.Set("k1", "v1")
	f1.Set("k2", "v2")
	if f1["k1"] != "v1" || f1["k2"] != "v2" {
		t.Fatal("Files_Set test failed")
	}

	if f1.Get("k1") != "v1" || f1.Get("k2") != "v2" {
		t.Error("Files_Get test failed")
	}

	f1.Del("k1")
	if f1["k1"] != "" || len(f1) != 1 {
		t.Error("Files_Del test failed")
	}

	f2 := make(sreq.Files)
	if err := json.Unmarshal([]byte(f1.String()), &f2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f2, f1) {
		t.Error("Files_String test failed")
	}
}
