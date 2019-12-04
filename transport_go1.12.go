// +build go1.12
// +build !go1.13

package sreq

import (
	"net"
	"net/http"
	"time"
)

// DefaultTransport returns an HTTP transport used by DefaultClient.
// It's a clone of http.DefaultTransport indeed.
func DefaultTransport() http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}
