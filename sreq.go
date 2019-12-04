package sreq

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"sync"
)

const (
	// Version of sreq.
	Version = "0.1.0"
)

var (
	bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

type (
	// Params is the same as map[string]string, used for query params.
	Params map[string]string

	// Headers is the same as map[string]string, used for request headers.
	Headers map[string]string

	// Form is the same as map[string]string, used for form-data.
	Form map[string]string

	// JSON is the same as map[string]interface{}, used for JSON payload.
	JSON map[string]interface{}

	// Files is the same as map[string]string, used for multipart-data.
	Files map[string]string
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
func (p Params) Get(key string) string {
	return p[key]
}

// Set sets a kv pair into a map.
func (p Params) Set(key string, value string) {
	p[key] = value
}

// Del deletes the value related to the given key from a map.
func (p Params) Del(key string) {
	delete(p, key)
}

// Clone returns a copy of p or nil if p is nil.
func (p Params) Clone() Params {
	if p == nil {
		return nil
	}

	p2 := make(Params, len(p))
	for k, v := range p {
		p2[k] = v
	}
	return p2
}

// Merge merges params to the copy of p and returns the merged Params.
func (p Params) Merge(params Params) Params {
	p2 := p.Clone()
	if p2 == nil {
		return params
	}

	for k, v := range params {
		p2[k] = v
	}
	return p2
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
func (h Headers) Get(key string) string {
	return h[key]
}

// Set sets a kv pair into a map.
func (h Headers) Set(key string, value string) {
	h[key] = value
}

// Del deletes the value related to the given key from a map.
func (h Headers) Del(key string) {
	delete(h, key)
}

// Clone returns a copy of h or nil if h is nil.
func (h Headers) Clone() Headers {
	if h == nil {
		return nil
	}

	h2 := make(Headers, len(h))
	for k, v := range h {
		h2[k] = v
	}
	return h2
}

// Merge merges headers to the copy of h and returns the merged Headers.
func (h Headers) Merge(headers Headers) Headers {
	h2 := h.Clone()
	if h2 == nil {
		return headers
	}

	for k, v := range headers {
		h2[k] = v
	}
	return h2
}

// String returns the JSON-encoded text representation of h.
func (h Headers) String() string {
	return toJSON(h)
}

// Get returns the value from a map by the given key.
func (f Form) Get(key string) string {
	return f[key]
}

// Set sets a kv pair into a map.
func (f Form) Set(key string, value string) {
	f[key] = value
}

// Del deletes the value related to the given key from a map.
func (f Form) Del(key string) {
	delete(f, key)
}

// Clone returns a copy of f or nil if f is nil.
func (f Form) Clone() Form {
	if f == nil {
		return nil
	}

	f2 := make(Form, len(f))
	for k, v := range f {
		f2[k] = v
	}
	return f2
}

// Merge merges form to the copy of f and returns the merged Form.
func (f Form) Merge(form Form) Form {
	f2 := f.Clone()
	if f2 == nil {
		return form
	}

	for k, v := range form {
		f2[k] = v
	}
	return f2
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

// Clone returns a copy of j or nil if j is nil.
func (j JSON) Clone() JSON {
	if j == nil {
		return nil
	}

	j2 := make(JSON, len(j))
	for k, v := range j {
		j2[k] = v
	}
	return j2
}

// Merge merges data to the copy of j and returns the merged JSON.
func (j JSON) Merge(data JSON) JSON {
	j2 := j.Clone()
	if j2 == nil {
		return data
	}

	for k, v := range data {
		j2[k] = v
	}
	return j2
}

// String returns the JSON-encoded text representation of j.
func (j JSON) String() string {
	return toJSON(j)
}

// Get returns the value from a map by the given key.
func (f Files) Get(key string) string {
	return f[key]
}

// Set sets a kv pair into a map.
func (f Files) Set(key string, value string) {
	f[key] = value
}

// Del deletes the value related to the given key from a map.
func (f Files) Del(key string) {
	delete(f, key)
}

// String returns the JSON-encoded text representation of f.
func (f Files) String() string {
	return toJSON(f)
}

func urlEncode(v map[string]string) string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		if sb.Len() > 0 {
			sb.WriteString("&")
		}

		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v[k])
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
