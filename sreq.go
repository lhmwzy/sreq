package sreq

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// Version of sreq.
	Version = "0.5.0"

	defaultUserAgent = "go-sreq/" + Version
)

var (
	bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

type (
	// Values maps a string key to an interface{} type value,
	// support string, int, []string, []int or []interface{} only with string and int.
	// Used for query parameters and form values.
	Values map[string]interface{}

	// Params is an alias of Values, used for for query parameters.
	Params = Values

	// Form is an alias of Values, used for form values.
	Form = Values

	// Headers maps a string key to an interface{} type value,
	// support string, int, []string, []int or []interface{} only with string and int.
	// Used for headers.
	Headers map[string]interface{}

	// JSON maps a string key to an interface{} type value, used for JSON payload.
	JSON map[string]interface{}

	// Files maps a string key to a *FileForm type value, used for files of multipart payload.
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

	// KV is the interface that defines a data type used by sreq in many cases.
	// The Keys method should return a slice of keys from a map.
	// The Get method should return a slice of values typed string associated with the given key.
	KV interface {
		Keys() []string
		Get(key string) []string
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

// Get gets the value associated with the given key, ignore unsupported data type.
func (v Values) Get(key string) []string {
	if v == nil {
		return nil
	}

	return filter(v[key])
}

// Set sets the key to value. It replaces any existing values.
func (v Values) Set(key string, value interface{}) {
	v[key] = value
}

// Del deletes the values associated with key.
func (v Values) Del(key string) {
	delete(v, key)
}

// Keys returns the keys of v.
func (v Values) Keys() []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}

// Encode encodes v into URL-unescaped form sorted by key.
func (v Values) Encode() string {
	var sb strings.Builder
	write(&sb, v, writeValues)
	return sb.String()
}

// String returns the text representation of v.
func (v Values) String() string {
	return v.Encode()
}

// Get gets the value associated with the given key, ignore unsupported data type.
func (h Headers) Get(key string) []string {
	if h == nil {
		return nil
	}

	return filter(h[key])
}

// Set sets the key to value. It replaces any existing values.
func (h Headers) Set(key string, value interface{}) {
	h[key] = value
}

// Del deletes the values associated with key.
func (h Headers) Del(key string) {
	delete(h, key)
}

// Keys returns the keys of h.
func (h Headers) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

// String returns the text representation of h.
func (h Headers) String() string {
	var sb strings.Builder
	write(&sb, h, writeHeaders)
	return sb.String()
}

// Get gets the value associated with the given key.
func (j JSON) Get(key string) interface{} {
	return j[key]
}

// Set sets the key to value. It replaces any existing values.
func (j JSON) Set(key string, value interface{}) {
	j[key] = value
}

// Del deletes the values associated with key.
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

// Set sets the key to value. It replaces any existing values.
func (f Files) Set(key string, value *FileForm) {
	f[key] = value
}

// Del deletes the values associated with key.
func (f Files) Del(key string) {
	delete(f, key)
}

// NewFileForm returns a *FileForm instance given a filename and its body.
func NewFileForm(filename string, body io.Reader) *FileForm {
	return &FileForm{
		Filename: filename,
		Body:     body,
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

// Read implements Reader interface.
func (ff *FileForm) Read(p []byte) (n int, err error) {
	if ff.Body == nil {
		return 0, io.EOF
	}
	return ff.Body.Read(p)
}

// Close implements Closer interface.
func (ff *FileForm) Close() error {
	if ff.Body == nil {
		return nil
	}

	rc, ok := ff.Body.(io.ReadCloser)
	if !ok {
		rc = ioutil.NopCloser(ff.Body)
	}
	return rc.Close()
}

// Open opens the named file and returns a *FileForm instance whose Filename is filename.
func Open(filename string) (*FileForm, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewFileForm(filename, file), nil
}

// MustOpen opens the named file and returns a *FileForm instance whose Filename is filename.
// If there is an error, it will panic.
func MustOpen(filename string) *FileForm {
	ff, err := Open(filename)
	if err != nil {
		panic(err)
	}

	return ff
}

func convertIntArray(v []int) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.Itoa(vv)
	}
	return vs
}

func convertStringIntArray(v []interface{}) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		switch vv := vv.(type) {
		case string:
			vs[i] = vv
		case int:
			vs[i] = strconv.Itoa(vv)
		}
	}
	return vs
}

func filter(v interface{}) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case int:
		return []string{strconv.Itoa(v)}
	case []string:
		return v
	case []int:
		return convertIntArray(v)
	case []interface{}:
		return convertStringIntArray(v)
	default:
		return nil
	}
}

func writeValues(sb *strings.Builder, k string, v []string) {
	for _, vs := range v {
		if sb.Len() > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(vs)
	}
}

func writeHeaders(sb *strings.Builder, k string, v []string) {
	for _, vs := range v {
		if sb.Len() > 0 {
			sb.WriteString("\r\n")
		}
		sb.WriteString(http.CanonicalHeaderKey(k))
		sb.WriteString(": ")
		sb.WriteString(vs)
	}
}

func write(sb *strings.Builder, v KV,
	callback func(sb *strings.Builder, k string, v []string)) {
	keys := v.Keys()
	sort.Strings(keys)

	for _, k := range keys {
		callback(sb, k, v.Get(k))
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
