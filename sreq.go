package sreq

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// Version of sreq.
	Version = "0.3.0"
)

var (
	bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

type (
	// Params is the same as map[string]interface{}, used for query params.
	Params map[string]interface{}

	// Headers is the same as map[string]interface{}, used for request headers.
	Headers map[string]interface{}

	// Form is the same as map[string]interface{}, used for form-data.
	Form map[string]interface{}

	// JSON is the same as map[string]interface{}, used for JSON payload.
	JSON map[string]interface{}

	// Files specifies files of multipart payload.
	Files map[string]*FileForm

	// FileForm specifies a file form.
	// If the Reader isn't an *os.File instance and you do not specify the FileName,
	// sreq will consider it as a form value.
	FileForm struct {
		Reader   io.Reader
		FileName string
		MIME     string
	}

	retry struct {
		attempts   int
		delay      time.Duration
		conditions []func(*Response) bool
	}
)

func acquireBuffer() *bytes.Buffer {
	return bufPool.Get().(*bytes.Buffer)
}

func releaseBuffer(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		bufPool.Put(buf)
	}
}

// Get returns the value from a map by the given key.
func (p Params) Get(key string) interface{} {
	return p[key]
}

// Set sets a kv pair into a map.
func (p Params) Set(key string, value interface{}) {
	p[key] = value
}

// Del deletes the value related to the given key from a map.
func (p Params) Del(key string) {
	delete(p, key)
}

// Encode encodes p into URL-unescaped form sorted by key.
func (p Params) Encode() string {
	return urlEncode(p)
}

// String returns the text representation of p.
func (p Params) String() string {
	return p.Encode()
}

// Get returns the value from a map by the given key.
func (h Headers) Get(key string) interface{} {
	return h[key]
}

// Set sets a kv pair into a map.
func (h Headers) Set(key string, value interface{}) {
	h[key] = value
}

// Del deletes the value related to the given key from a map.
func (h Headers) Del(key string) {
	delete(h, key)
}

// String returns the JSON-encoded text representation of h.
func (h Headers) String() string {
	return toJSON(h)
}

// Get returns the value from a map by the given key.
func (f Form) Get(key string) interface{} {
	return f[key]
}

// Set sets a kv pair into a map.
func (f Form) Set(key string, value interface{}) {
	f[key] = value
}

// Del deletes the value related to the given key from a map.
func (f Form) Del(key string) {
	delete(f, key)
}

// Encode encodes f into URL-unescaped form sorted by key.
func (f Form) Encode() string {
	return urlEncode(f)
}

// String returns the text representation of f.
func (f Form) String() string {
	return f.Encode()
}

// Get returns the value from a map by the given key.
func (j JSON) Get(key string) interface{} {
	return j[key]
}

// Set sets a kv pair into a map.
func (j JSON) Set(key string, value interface{}) {
	j[key] = value
}

// Del deletes the value related to the given key from a map.
func (j JSON) Del(key string) {
	delete(j, key)
}

// String returns the JSON-encoded text representation of j.
func (j JSON) String() string {
	return toJSON(j)
}

// Get returns the value from a map by the given key.
func (f Files) Get(key string) *FileForm {
	return f[key]
}

// Set sets a kv pair into a map.
func (f Files) Set(key string, value *FileForm) {
	f[key] = value
}

// Del deletes the value related to the given key from a map.
func (f Files) Del(key string) {
	delete(f, key)
}

// MustOpen opens the named file for reading.
// If there is an error, it will panic.
func MustOpen(filename string) *os.File {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	return file
}

func addPair(sb *strings.Builder, k string, v string) {
	if sb.Len() > 0 {
		sb.WriteString("&")
	}
	sb.WriteString(k)
	sb.WriteString("=")
	sb.WriteString(v)
}

func setStringArray(sb *strings.Builder, k string, v []string) {
	for _, vs := range v {
		addPair(sb, k, vs)
	}
}

func setIntArray(sb *strings.Builder, k string, v []int) {
	for _, vs := range v {
		addPair(sb, k, strconv.Itoa(vs))
	}
}

func setStringIntArray(sb *strings.Builder, k string, v []interface{}) {
	for _, vs := range v {
		switch vs := vs.(type) {
		case string:
			addPair(sb, k, vs)
		case int:
			addPair(sb, k, strconv.Itoa(vs))
		}
	}
}

func urlEncode(v map[string]interface{}) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		switch v := v[k].(type) {
		case string:
			addPair(&sb, k, v)
		case int:
			addPair(&sb, k, strconv.Itoa(v))
		case []string:
			setStringArray(&sb, k, v)
		case []int:
			setIntArray(&sb, k, v)
		case []interface{}:
			setStringIntArray(&sb, k, v)
		}
	}

	return sb.String()
}

func toJSON(data interface{}) string {
	b, err := jsonMarshal(data, "", "\t", false)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func jsonMarshal(v interface{}, prefix string, indent string, escapeHTML bool) ([]byte, error) {
	buf := acquireBuffer()
	defer releaseBuffer(buf)

	encoder := json.NewEncoder(buf)
	encoder.SetIndent(prefix, indent)
	encoder.SetEscapeHTML(escapeHTML)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}
