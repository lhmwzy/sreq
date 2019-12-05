package sreq_test

import (
	"strings"
	"testing"

	"github.com/winterssy/sreq"
)

func TestClientError(t *testing.T) {
	_, err := sreq.
		New().
		SetTransport(nil).
		DisableVerify().
		Raw()
	if err == nil {
		t.Fatal("ClientError test failed")
	}

	cErr, ok := err.(*sreq.ClientError)
	if !ok || !strings.HasPrefix(cErr.Error(), "sreq [Client]") {
		t.Error("ClientError test failed")
	}

	if cErr.Unwrap() != sreq.ErrUnexpectedTransport {
		t.Error("ClientError test failed")
	}
}

func TestRequestError(t *testing.T) {
	_, err := sreq.
		NewRequest(sreq.MethodGet, "https://www.google.com").
		SetContext(nil).
		Raw()
	if err == nil {
		t.Fatal("RequestError test failed")
	}

	reqErr, ok := err.(*sreq.RequestError)
	if !ok || !strings.HasPrefix(reqErr.Error(), "sreq [Request]") {
		t.Error("RequestError test failed")
	}

	if reqErr.Unwrap() != sreq.ErrNilContext {
		t.Error("RequestError test failed")
	}
}
