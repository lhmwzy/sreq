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

// Get gets the value associated with the given key.
func (v Values) Get(key string) interface{} {
	return v[key]
}

// Set sets the key to value. It replaces any existing values.
func (v Values) Set(key string, value interface{}) {
	v[key] = value
}

// Del deletes the values associated with key.
func (v Values) Del(key string) {
	delete(v, key)
}

func addValuePair(sb *strings.Builder, k string, v string) {
	if sb.Len() > 0 {
		sb.WriteString("&")
	}
	sb.WriteString(k)
	sb.WriteString("=")
	sb.WriteString(v)
}

// Encode encodes v into URL-unescaped form sorted by key.
func (v Values) Encode() string {
	var sb strings.Builder
	return output(&sb, v, addValuePair)
}

// String returns the text representation of v.
func (v Values) String() string {
	return v.Encode()
}

// Get gets the value associated with the given key.
func (h Headers) Get(key string) interface{} {
	return h[key]
}

// Set sets the key to value. It replaces any existing values.
func (h Headers) Set(key string, value interface{}) {
	h[key] = value
}

// Del deletes the values associated with key.
func (h Headers) Del(key string) {
	delete(h, key)
}

func addHeadersPair(sb *strings.Builder, k string, v string) {
	if sb.Len() > 0 {
		sb.WriteString("\r\n")
	}
	sb.WriteString(http.CanonicalHeaderKey(k))
	sb.WriteString(": ")
	sb.WriteString(v)
}

// String returns the text representation of h.
func (h Headers) String() string {
	var sb strings.Builder
	return output(&sb, h, addHeadersPair)
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

func output(sb *strings.Builder, v map[string]interface{},
	callback func(*strings.Builder, string, string)) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		switch v := v[k].(type) {
		case string:
			callback(sb, k, v)
		case int:
			callback(sb, k, strconv.Itoa(v))
		case []string:
			addStringArray(sb, k, v, callback)
		case []int:
			addIntArray(sb, k, v, callback)
		case []interface{}:
			addStringIntArray(sb, k, v, callback)
		}
	}

	return sb.String()
}

func addStringArray(sb *strings.Builder, k string, v []string,
	callback func(*strings.Builder, string, string)) {
	for _, vs := range v {
		callback(sb, k, vs)
	}
}

func addIntArray(sb *strings.Builder, k string, v []int,
	callback func(*strings.Builder, string, string)) {
	for _, vs := range v {
		callback(sb, k, strconv.Itoa(vs))
	}
}

func addStringIntArray(sb *strings.Builder, k string, v []interface{},
	callback func(*strings.Builder, string, string)) {
	for _, vs := range v {
		switch vs := vs.(type) {
		case string:
			callback(sb, k, vs)
		case int:
			callback(sb, k, strconv.Itoa(vs))
		}
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
