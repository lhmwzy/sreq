package sreq_test

import (
	"context"
	"encoding/json"
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

const (
	testString = "2019"
	testInt    = 2019
	wantValues = "int=2019&intArray=10086&intArray=10010&" +
		"string=2019&stringArray=10086&stringArray=10010&" +
		"stringIntArray=10086&stringIntArray=10010"
	wantHeaders = "Int: 2019\r\nInt-Array: 10086\r\nInt-Array: 10010\r\n" +
		"String: 2019\r\nString-Array: 10086\r\nString-Array: 10010\r\n" +
		"String-Int-Array: 10086\r\nString-Int-Array: 10010"
)

var (
	testStringArray    = []string{"10086", "10010"}
	testIntArray       = []int{10086, 10010}
	testStringIntArray = []interface{}{"10086", 10010}
	testValues         = sreq.Values{
		"string":         testString,
		"int":            testInt,
		"stringArray":    testStringArray,
		"intArray":       testIntArray,
		"stringIntArray": testStringIntArray,
	}
	testHeaders = sreq.Headers{
		"string":           testString,
		"int":              testInt,
		"string-array":     testStringArray,
		"int-array":        testIntArray,
		"string-int-array": testStringIntArray,
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
	v.Set("string", testString)
	v.Set("int", testInt)
	v.Set("stringArray", testStringArray)
	v.Set("intArray", testIntArray)
	v.Set("stringIntArray", testStringIntArray)
	if len(v) != 5 {
		t.Fatal("Values_Set test failed")
	}

	if !reflect.DeepEqual(v.Get("stringArray"), testStringArray) ||
		!reflect.DeepEqual(v.Get("intArray"), testStringArray) ||
		!reflect.DeepEqual(v.Get("stringIntArray"), testStringArray) {
		t.Error("Values_Get test failed")
	}

	if got := v.String(); got != wantValues {
		t.Errorf("Values_String got: %q, want: %q", got, wantValues)
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
	want := "limit=100&offset=0&q=Go语言"
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
	h.Set("string", testString)
	h.Set("int", testInt)
	h.Set("string-array", testStringArray)
	h.Set("int-array", testIntArray)
	h.Set("string-int-array", testStringIntArray)
	if len(h) != 5 {
		t.Fatal("Headers_Set test failed")
	}

	if !reflect.DeepEqual(h.Get("string-array"), testStringArray) ||
		!reflect.DeepEqual(h.Get("int-array"), testStringArray) ||
		!reflect.DeepEqual(h.Get("string-int-array"), testStringArray) {
		t.Error("Headers_Get test failed")
	}

	if got := h.String(); got != wantHeaders {
		t.Errorf("Headers_String got: %q, want: %q", got, wantHeaders)
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

func TestH(t *testing.T) {
	var (
		boolSlice            = []bool{true, false}
		stringVal            = "hello world"
		stringSlice          = []string{"hello", "world"}
		float64Val           = 3.14159
		float64Slice         = []float64{3.14159, 3.14160}
		float32Val   float32 = 3.14
		float32Slice         = []float32{3.14, 3.15}
		intVal               = -314
		intSlice             = []int{-314, -315}
		int32Val     int32   = -314159
		int32Slice           = []int32{-314159, -314160}
		int64Val     int64   = -31415926535
		int64Slice           = []int64{-31415926535, -31415926536}
		uintVal      uint    = 314
		uintSlice            = []uint{314, 315}
		uint32Val    uint32  = 314159
		uint32Slice          = []uint32{314159, 314160}
		uint64Val    uint64  = 31415926535
		uint64Slice          = []uint64{31415926535, 31415926536}
	)

	m := map[string]interface{}{
		"bool":         true,
		"boolSlice":    boolSlice,
		"string":       stringVal,
		"stringSlice":  stringSlice,
		"float64":      float64Val,
		"float64Slice": float64Slice,
		"float32":      float32Val,
		"float32Slice": float32Slice,
		"int":          intVal,
		"intSlice":     intSlice,
		"int32":        int32Val,
		"int32Slice":   int32Slice,
		"int64":        int64Val,
		"int64Slice":   int64Slice,
		"uint":         uintVal,
		"uintSlice":    uintSlice,
		"uint32":       uint32Val,
		"uint32Slice":  uint32Slice,
		"uint64":       uint64Val,
		"uint64Slice":  uint64Slice,
	}
	enc, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	var h sreq.H
	if h.Get("key") != nil || len(h.GetSlice("slice")) != 0 ||
		h.GetBool("bool") || h.GetString("string") != "" ||
		h.GetFloat64("float64") != 0 || h.GetH("key") != nil {
		t.Error("H test failed")
	}

	err = json.Unmarshal(enc, &h)
	if err != nil {
		t.Fatal(err)
	}

	if h.Get("float64") != float64Val || !h.GetBool("bool") ||
		!reflect.DeepEqual(h.GetBoolSlice("boolSlice"), boolSlice) ||
		h.GetString("string") != stringVal ||
		!reflect.DeepEqual(h.GetStringSlice("stringSlice"), stringSlice) ||
		h.GetFloat64("float64") != float64Val ||
		!reflect.DeepEqual(h.GetFloat64Slice("float64Slice"), float64Slice) ||
		h.GetFloat32("float32") != float32Val ||
		!reflect.DeepEqual(h.GetFloat32Slice("float32Slice"), float32Slice) ||
		h.GetInt("int") != intVal ||
		!reflect.DeepEqual(h.GetIntSlice("intSlice"), intSlice) ||
		h.GetInt32("int32") != int32Val ||
		!reflect.DeepEqual(h.GetInt32Slice("int32Slice"), int32Slice) ||
		h.GetInt64("int64") != int64Val ||
		!reflect.DeepEqual(h.GetInt64Slice("int64Slice"), int64Slice) ||
		h.GetUint("uint") != uintVal ||
		!reflect.DeepEqual(h.GetUintSlice("uintSlice"), uintSlice) ||
		h.GetUint32("uint32") != uint32Val ||
		!reflect.DeepEqual(h.GetUint32Slice("uint32Slice"), uint32Slice) ||
		h.GetUint64("uint64") != uint64Val ||
		!reflect.DeepEqual(h.GetUint64Slice("uint64Slice"), uint64Slice) {
		t.Error("H test failed")
	}

	h = sreq.H{
		"msg": "hello world",
	}
	want := "{\n\t\"msg\": \"hello world\"\n}\n"
	if got := h.String(); got != want {
		t.Errorf("H_string got: %q, want: %q", got, want)
	}
}
