//go:build sync

package sync

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"zotregistry.dev/zot/v2/pkg/log"
)

type testRoundTripper func(*http.Request) (*http.Response, error)

func (trt testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return trt(req)
}

func TestHTTP2FramingError(t *testing.T) {
	tests := map[string]bool{
		"malformed HTTP response":                   true,
		"stream error: stream ID 1; INTERNAL_ERROR": true,
		"PROTOCOL_ERROR":                            true,
		"unexpected EOF":                            false,
		"dial tcp: i/o timeout":                     false,
	}

	for message, expected := range tests {
		if got := isHTTP2FramingError(errors.New(message)); got != expected {
			t.Fatalf("unexpected classification for %q: got %v, expected %v", message, got, expected)
		}
	}
}

func TestHTTP2FallbackTransport(t *testing.T) {
	logger := log.NewTestLogger()

	req, err := http.NewRequest(http.MethodGet, "https://index.docker.io/v2/library/alpine/blobs/sha256:abc", nil)
	if err != nil {
		t.Fatal(err)
	}

	primaryCount, fallbackCount := 0, 0

	transport := &http2FallbackTransport{
		primary: testRoundTripper(func(*http.Request) (*http.Response, error) {
			primaryCount++
			return nil, errors.New("malformed HTTP response \"HTTP/1.1 200 OK\\r\\n...\"")
		}),
		fallback: testRoundTripper(func(_ *http.Request) (*http.Response, error) {
			fallbackCount++
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
		log: logger,
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
	}

	if primaryCount != 1 || fallbackCount != 1 {
		t.Fatalf("unexpected transport call count: primary=%d fallback=%d", primaryCount, fallbackCount)
	}
}

func TestHTTP2FallbackTransport_NoRetry(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://index.docker.io/v2/library/alpine/blobs/sha256:def", nil)
	if err != nil {
		t.Fatal(err)
	}

	primaryCount, fallbackCount := 0, 0

	transport := &http2FallbackTransport{
		primary: testRoundTripper(func(*http.Request) (*http.Response, error) {
			primaryCount++
			return nil, errors.New("unexpected EOF")
		}),
		fallback: testRoundTripper(func(*http.Request) (*http.Response, error) {
			fallbackCount++
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header)}, nil
		}),
		log: log.NewTestLogger(),
	}

	_, err = transport.RoundTrip(req)
	if err == nil {
		t.Fatalf("expected error")
	}

	if primaryCount != 1 || fallbackCount != 0 {
		t.Fatalf("expected only primary round trip on non-framing errors, got primary=%d fallback=%d", primaryCount, fallbackCount)
	}
}
