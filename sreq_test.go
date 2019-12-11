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

var (
	testStringArray = []string{"10086", "10010", "10000"}
	testValues      = sreq.Values{
		"string":         "2019",
		"int":            2019,
		"stringArray":    testStringArray,
		"intArray":       []int{10086, 10010, 10000},
		"stringIntArray": []interface{}{"10086", 10010, 10000},
	}
	testHeaders = sreq.Headers{
		"string":           "2019",
		"int":              2019,
		"string-array":     testStringArray,
		"int-array":        []int{10086, 10010, 10000},
		"string-int-array": []interface{}{"10086", 10010, 10000},
	}
)

type (
	testStruct struct {
		String         string   `json:"string"`
		Int            string   `json:"int"`
		StringArray    []string `json:"stringArray"`
		IntArray       []string `json:"intArray"`
		StringIntArray []string `json:"stringIntArray"`
	}
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
	v.Set("string", "2019")
	v.Set("int", 2019)
	v.Set("stringArray", testStringArray)
	v.Set("intArray", []int{10086, 10010, 10000})
	v.Set("stringIntArray", []interface{}{"10086", 10010, 10000})
	if len(v) != 5 {
		t.Fatal("Values_Set test failed")
	}

	if !reflect.DeepEqual(v.Get("stringArray"), testStringArray) ||
		!reflect.DeepEqual(v.Get("intArray"), testStringArray) ||
		!reflect.DeepEqual(v.Get("stringIntArray"), testStringArray) {
		t.Error("Values_Get test failed")
	}

	want := "int=2019&intArray=10086&intArray=10010&intArray=10000&" +
		"string=2019&stringArray=10086&stringArray=10010&stringArray=10000&" +
		"stringIntArray=10086&stringIntArray=10010&stringIntArray=10000"
	if got := v.String(); got != want {
		t.Errorf("Values_String got: %q, want: %q", got, want)
	}

	v.Del("string")
	if len(v.Get("string")) != 0 || len(v) != 4 {
		t.Error("Values_Del test failed")
	}

	v = sreq.Values{
		"q":      "Go语言",
		"offset": 0,
		"limit":  100,
	}
	want = "limit=100&offset=0&q=Go语言"
	if got := v.Encode(); got != want {
		t.Errorf("Values_Encode got: %q, want: %q", got, want)
	}
}

func TestHeaders(t *testing.T) {
	var h sreq.Headers
	if h.Get("key") != nil {
		t.Error("Headers_Get test failed")
	}

	h = make(sreq.Headers)
	h.Set("string", "2019")
	h.Set("int", 2019)
	h.Set("string-array", testStringArray)
	h.Set("int-array", []int{10086, 10010, 10000})
	h.Set("string-int-array", []interface{}{"10086", 10010, 10000})
	if len(h) != 5 {
		t.Fatal("Headers_Set test failed")
	}

	if !reflect.DeepEqual(h.Get("string-array"), testStringArray) ||
		!reflect.DeepEqual(h.Get("int-array"), testStringArray) ||
		!reflect.DeepEqual(h.Get("string-int-array"), testStringArray) {
		t.Error("Headers_Get test failed")
	}

	want := "Int: 2019\r\nInt-Array: 10086\r\nInt-Array: 10010\r\nInt-Array: 10000\r\n" +
		"String: 2019\r\nString-Array: 10086\r\nString-Array: 10010\r\nString-Array: 10000\r\n" +
		"String-Int-Array: 10086\r\nString-Int-Array: 10010\r\nString-Int-Array: 10000"
	if got := h.String(); got != want {
		t.Errorf("Headers_String got: %q, want: %q", got, want)
	}

	h.Del("string")
	if len(h.Get("string")) != 0 || len(h) != 4 {
		t.Error("Headers_Del test failed")
	}
}

func TestFiles(t *testing.T) {
	var f sreq.Files
	if f.Get("key") != nil {
		t.Error("Files_Get test failed")
	}

	f = make(sreq.Files)
	f.Set("file1", &sreq.File{
		Body: &os.File{},
		MIME: "image/png",
	})
	f.Set("file2", &sreq.File{
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

	ff = &sreq.File{
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
