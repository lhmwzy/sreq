package sreq

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/text/encoding"
)

type (
	// Response wraps the raw HTTP response.
	Response struct {
		RawResponse *http.Response
		Err         error

		body []byte
	}

	// ResponseInterceptor specifies a response interceptor.
	ResponseInterceptor func(*Response) error
)

// Raw returns the raw HTTP response.
func (resp *Response) Raw() (*http.Response, error) {
	return resp.RawResponse, resp.Err
}

// Content decodes the HTTP response body to bytes.
func (resp *Response) Content() ([]byte, error) {
	if resp.Err != nil || resp.body != nil {
		return resp.body, resp.Err
	}
	defer resp.RawResponse.Body.Close()

	var err error
	resp.body, err = ioutil.ReadAll(resp.RawResponse.Body)
	return resp.body, err
}

// Text decodes the HTTP response body and returns the text representation of its raw data
// given an optional charset encoding.
func (resp *Response) Text(e ...encoding.Encoding) (string, error) {
	b, err := resp.Content()
	if err != nil || len(e) == 0 {
		return string(b), err
	}

	b, err = e[0].NewDecoder().Bytes(b)
	return string(b), err
}

// JSON decodes the HTTP response body and unmarshals its JSON-encoded data into v.
func (resp *Response) JSON(v interface{}) error {
	if resp.Err != nil {
		return resp.Err
	}

	if resp.body != nil {
		return json.Unmarshal(resp.body, v)
	}

	buf := acquireBuffer()
	tee := io.TeeReader(resp.RawResponse.Body, buf)
	defer func() {
		resp.RawResponse.Body.Close()
		resp.body = buf.Bytes()
		releaseBuffer(buf)
	}()

	return json.NewDecoder(tee).Decode(v)
}

// H decodes the HTTP response body and unmarshals its JSON-encoded data into a sreq.H instance.
func (resp *Response) H() (H, error) {
	h := make(H)
	return h, resp.JSON(&h)
}

// XML decodes the HTTP response body and unmarshals its XML-encoded data into v.
func (resp *Response) XML(v interface{}) error {
	if resp.Err != nil {
		return resp.Err
	}

	if resp.body != nil {
		return xml.Unmarshal(resp.body, v)
	}

	buf := acquireBuffer()
	tee := io.TeeReader(resp.RawResponse.Body, buf)
	defer func() {
		resp.RawResponse.Body.Close()
		resp.body = buf.Bytes()
		releaseBuffer(buf)
	}()

	return xml.NewDecoder(tee).Decode(v)
}

// Cookies returns the HTTP response cookies.
func (resp *Response) Cookies() ([]*http.Cookie, error) {
	if resp.Err != nil {
		return nil, resp.Err
	}

	cookies := resp.RawResponse.Cookies()
	if len(cookies) == 0 {
		return nil, ErrResponseCookiesNotPresent
	}

	return cookies, nil
}

// Cookie returns the HTTP response named cookie.
func (resp *Response) Cookie(name string) (*http.Cookie, error) {
	cookies, err := resp.Cookies()
	if err != nil {
		return nil, err
	}

	for _, c := range cookies {
		if c.Name == name {
			return c, nil
		}
	}

	return nil, ErrResponseNamedCookieNotPresent
}

// EnsureStatusOk ensures the HTTP response's status code must be 200.
func (resp *Response) EnsureStatusOk() *Response {
	return resp.EnsureStatus(http.StatusOK)
}

// EnsureStatus2xx ensures the HTTP response's status code must be 2xx.
func (resp *Response) EnsureStatus2xx() *Response {
	if resp.Err != nil {
		return resp
	}

	if resp.RawResponse.StatusCode/100 != 2 {
		resp.Err = fmt.Errorf("sreq: bad status: %d", resp.RawResponse.StatusCode)
	}
	return resp
}

// EnsureStatus ensures the HTTP response's status code must be the code parameter.
func (resp *Response) EnsureStatus(code int) *Response {
	if resp.Err != nil {
		return resp
	}

	if resp.RawResponse.StatusCode != code {
		resp.Err = fmt.Errorf("sreq: bad status: %d", resp.RawResponse.StatusCode)
	}
	return resp
}

// Save saves the HTTP response into a file.
// Notes: Save won't make the HTTP response body reused.
func (resp *Response) Save(filename string, perm os.FileMode) error {
	if resp.Err != nil {
		return resp.Err
	}

	if resp.body != nil {
		return ioutil.WriteFile(filename, resp.body, perm)
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()
	defer resp.RawResponse.Body.Close()

	_, err = io.Copy(file, resp.RawResponse.Body)
	return err
}

// Verbose makes the HTTP request and its response more talkative.
// It's similar to "curl -v", used for debug.
// Notes: Verbose won't make the HTTP response body reused.
func (resp *Response) Verbose(w io.Writer) error {
	if resp.Err != nil {
		return resp.Err
	}

	rawRequest := resp.RawResponse.Request
	fmt.Fprintf(w, "> %s %s %s\r\n", rawRequest.Method, rawRequest.URL.RequestURI(), rawRequest.Proto)
	fmt.Fprintf(w, "> Host: %s\r\n", rawRequest.URL.Host)
	for k := range rawRequest.Header {
		fmt.Fprintf(w, "> %s: %s\r\n", k, rawRequest.Header.Get(k))
	}
	fmt.Fprint(w, ">\r\n")

	if rawRequest.GetBody != nil && rawRequest.ContentLength != 0 {
		rc, err := rawRequest.GetBody()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}

		fmt.Fprint(w, "\r\n")
	}

	rawResponse := resp.RawResponse
	fmt.Fprintf(w, "< %s %s\r\n", rawResponse.Proto, rawResponse.Status)
	for k := range rawResponse.Header {
		fmt.Fprintf(w, "< %s: %s\r\n", k, rawResponse.Header.Get(k))
	}
	fmt.Fprint(w, "<\r\n")

	if resp.RawResponse.ContentLength == 0 {
		return nil
	}

	if resp.body != nil {
		fmt.Fprintf(w, "%s\r\n", string(resp.body))
		return nil
	}

	defer rawResponse.Body.Close()
	_, err := io.Copy(w, rawResponse.Body)
	if err != nil {
		return err
	}

	fmt.Fprint(w, "\r\n")
	return nil
}
