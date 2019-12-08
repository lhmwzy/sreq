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
	Version = "0.3.7"

	defaultUserAgent = "go-sreq/" + Version
)

var (
	bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

type (
	// Values is the same as map[string]interface{}, used for query params and form.
	Values map[string]interface{}

	// Params is an alias of Values.
	Params = Values

	// Form is an alias of Values.
	Form = Values

	// Headers is the same as map[string]interface{}, used for request headers.
	Headers map[string]interface{}

	// JSON is the same as map[string]interface{}, used for JSON payload.
	JSON map[string]interface{}

	// Files specifies files of multipart payload.
	Files map[string]*FileForm

	// FileForm specifies a file form.
	// To upload a file you must specify its Filename field,
	// otherwise sreq will consider it as a origin form, but not a file form.
	// If you don't specify the MIME field, sreq will detect automatically using http.DetectContentType.
	FileForm struct {
		Filename string
		Body     io.Reader
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

// Get returns the value related to the given key from a map.
func (v Values) Get(key string) interface{} {
	return v[key]
}

// Set sets a kv pair into a map.
func (v Values) Set(key string, value interface{}) {
	v[key] = value
}

// Del deletes the value related to the given key from a map.
func (v Values) Del(key string) {
	delete(v, key)
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

// Encode encodes v into URL-unescaped form sorted by key.
func (v Values) Encode() string {
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

// String returns the text representation of v.
func (v Values) String() string {
	return v.Encode()
}

// Get returns the value related to the given key from a map.
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

// Get returns the value related to the given key from a map.
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

// Get returns the value related to the given key from a map.
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

// NewFileForm returns a *FileForm instance given a body.
func NewFileForm(body io.Reader) *FileForm {
	return &FileForm{
		Body: body,
	}
}

// SetFilename sets Filename field value of ff.
func (ff *FileForm) SetFilename(filename string) *FileForm {
	ff.Filename = filename
	return ff
}

// SetMIME sets MIME field value of ff.
func (ff *FileForm) SetMIME(mime string) *FileForm {
	ff.MIME = mime
	return ff
}

// MustOpen opens the named file and returns a *FileForm instance whose Filename is filename.
// If there is an error, it will panic.
func MustOpen(filename string) *FileForm {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	return &FileForm{
		Filename: filename,
		Body:     file,
	}
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
