package sreq_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/winterssy/sreq"
)

// 用于测试 Client 是否复用连接
func printLocalDial(ctx context.Context, network, addr string) (net.Conn, error) {
	dial := net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	conn, err := dial.DialContext(ctx, network, addr)
	if err != nil {
		return conn, err
	}

	fmt.Printf("network connected at %s\n", conn.LocalAddr().String())
	return conn, err
}

func TestValues(t *testing.T) {
	v := make(sreq.Values)

	v.Set("k1", "v1")
	v.Set("k2", "v2")
	v.Set("k3", "v3")
	if v["k1"] != "v1" || v["k2"] != "v2" || v["k3"] != "v3" {
		t.Fatal("Values_Set test failed")
	}

	if v.Get("k1") != "v1" || v.Get("k2") != "v2" || v.Get("k3") != "v3" {
		t.Error("Values_Get test failed")
	}

	v.Del("k1")
	if v["k1"] != nil || len(v) != 2 {
		t.Error("Values_Del test failed")
	}

	want := "k2=v2&k3=v3"
	if got := v.String(); got != want {
		t.Errorf("Values_String got: %s, want: %s", got, want)
	}

	v = sreq.Params{
		"q":      "Go语言",
		"offset": 0,
		"limit":  100,
	}
	want = "limit=100&offset=0&q=Go语言"
	if got := v.Encode(); got != want {
		t.Errorf("Values_Encode got: %s, want: %s", got, want)
	}

	v = sreq.Params{
		"string":         "2019",
		"int":            2019,
		"stringArray":    []string{"10086", "10010"},
		"intArray":       []int{10086, 10010},
		"stringIntArray": []interface{}{"10086", 10010},
	}
	want = "int=2019&intArray=10086&intArray=10010&" +
		"string=2019&stringArray=10086&stringArray=10010&" +
		"stringIntArray=10086&stringIntArray=10010"
	if got := v.Encode(); got != want {
		t.Errorf("Values_Encode got: %s, want: %s", got, want)
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
		Body: &os.File{},
		MIME: "image/png",
	})
	f.Set("file2", &sreq.FileForm{
		Body: strings.NewReader("hello world"),
		MIME: "text/plain",
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
