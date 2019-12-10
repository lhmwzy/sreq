package sreq_test

import (
	"context"
	"fmt"
	"io/ioutil"
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
	var v sreq.Values
	if v.Get("key") != nil {
		t.Error("Values_Get test failed")
	}

	v = make(sreq.Values)
	stringArray := []string{"10086", "10010", "10000"}
	v.Set("string", "2019")
	v.Set("int", 2019)
	v.Set("stringArray", stringArray)
	v.Set("intArray", []int{10086, 10010, 10000})
	v.Set("stringIntArray", []interface{}{"10086", 10010, 10000})
	if len(v) != 5 {
		t.Fatal("Values_Set test failed")
	}

	if !reflect.DeepEqual(v.Get("stringArray"), stringArray) ||
		!reflect.DeepEqual(v.Get("intArray"), stringArray) ||
		!reflect.DeepEqual(v.Get("stringIntArray"), stringArray) {
		t.Error("Values_Get test failed")
	}

	v.Del("string")
	if len(v.Get("string")) != 0 || len(v) != 4 {
		t.Error("Values_Del test failed")
	}

	v = sreq.Params{
		"q":      "Go语言",
		"offset": 0,
		"limit":  100,
	}
	want := "limit=100&offset=0&q=Go语言"
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
	if got := v.String(); got != want {
		t.Errorf("Values_Encode got: %s, want: %s", got, want)
	}
}

func TestHeaders(t *testing.T) {
	var h sreq.Headers
	if h.Get("key") != nil {
		t.Error("Headers_Get test failed")
	}

	h = make(sreq.Headers)
	stringArray := []string{"10086", "10010", "10000"}
	h.Set("string", "2019")
	h.Set("int", 2019)
	h.Set("stringArray", stringArray)
	h.Set("intArray", []int{10086, 10010, 10000})
	h.Set("stringIntArray", []interface{}{"10086", 10010, 10000})
	if len(h) != 5 {
		t.Fatal("Headers_Set test failed")
	}

	if !reflect.DeepEqual(h.Get("stringArray"), stringArray) ||
		!reflect.DeepEqual(h.Get("intArray"), stringArray) ||
		!reflect.DeepEqual(h.Get("stringIntArray"), stringArray) {
		t.Error("Headers_Get test failed")
	}

	h.Del("string")
	if len(h.Get("string")) != 0 || len(h) != 4 {
		t.Error("Headers_Del test failed")
	}

	h = sreq.Headers{
		"string":           "2019",
		"int":              2019,
		"string-array":     []string{"10086", "10010"},
		"int-array":        []int{10086, 10010},
		"string-int-array": []interface{}{"10086", 10010},
	}
	want := "Int: 2019\r\nInt-Array: 10086\r\nInt-Array: 10010\r\n" +
		"String: 2019\r\nString-Array: 10086\r\nString-Array: 10010\r\n" +
		"String-Int-Array: 10086\r\nString-Int-Array: 10010"
	if got := h.String(); got != want {
		t.Errorf("Headers_String got: %s, want: %s", got, want)
	}
}

func TestJSON(t *testing.T) {
	var j sreq.JSON
	if j.Get("key") != nil {
		t.Error("JSON_Get test failed")
	}

	j = make(sreq.JSON)
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
	var f sreq.Files
	if f.Get("key") != nil {
		t.Error("Files_Get test failed")
	}

	f = make(sreq.Files)
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

func TestOpen(t *testing.T) {
	const (
		fileExist    = "./testdata/testfile1.txt"
		fileNotExist = "./testdata/file_not_exist.txt"
	)

	ff, err := sreq.Open(fileExist)
	if err != nil {
		t.Fatal(err)
	}
	if err = ff.Close(); err != nil {
		t.Error(err)
	}

	_, err = sreq.Open(fileNotExist)
	if err == nil {
		t.Error("Open test failed")
	}

	ff = &sreq.FileForm{
		Body: nil,
	}
	_, err = ioutil.ReadAll(ff)
	if err != nil {
		t.Error(err)
	}
	err = ff.Close()
	if err != nil {
		t.Error(err)
	}
}
