package util

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func ReadBody(reader io.ReadCloser) ([]byte, error) {
	if body, err := ioutil.ReadAll(reader); err == nil {
		err = reader.Close()
		return body, err
	} else {
		return nil, err
	}
}

func ReadHTTPBody(reader io.ReadCloser, encoding string) ([]byte, error) {
	var body []byte
	var err error

	if body, err = ReadBody(reader); err != nil {
		return nil, err
	}

	if encoding == "gzip" {
		var zr *gzip.Reader
		if zr, err = gzip.NewReader(bytes.NewReader(body)); err == nil {
			body, err = ioutil.ReadAll(zr)
			zr.Close()
		}
	} else if encoding == "deflate" {
		var zr io.ReadCloser
		if zr = flate.NewReader(bytes.NewReader(body)); zr != nil {
			body, err = ioutil.ReadAll(zr)
			zr.Close()
		}
	} else if len(encoding) > 0 {
		err = errors.New("unsupport encoding:" + encoding)
	}

	return body, err
}

// scheme derives the request scheme used on the initial
// request first from headers and then from the connection
// using the following heuristic:
//
// If either X-Forwarded-Proto or Forwarded is set then use
// its value to set the other header. If both headers are
// set do not modify the protocol. If none are set derive
// the protocol from the connection.
func ParseHTTPScheme(r *http.Request) string {
	xfp := r.Header.Get("X-Forwarded-Proto")
	fwd := r.Header.Get("Forwarded")
	switch {
	case xfp != "" && fwd == "":
		return xfp

	case fwd != "" && xfp == "":
		p := strings.SplitAfterN(fwd, "proto=", 2)
		if len(p) == 1 {
			break
		}
		n := strings.IndexRune(p[1], ';')
		if n >= 0 {
			return p[1][:n]
		}
		return p[1]
	}

	ws := r.Header.Get("Upgrade") == "websocket"
	switch {
	case ws && r.TLS != nil:
		return "wss"
	case ws && r.TLS == nil:
		return "ws"
	case r.TLS != nil:
		return "https"
	default:
		return "http"
	}
}

func ParseHTTPPort(r *http.Request) string {
	if r == nil {
		return ""
	}
	n := strings.Index(r.Host, ":")
	if n > 0 && n < len(r.Host)-1 {
		return r.Host[n+1:]
	}
	if r.TLS != nil {
		return "443"
	}
	return "80"
}

const (
	ChromeAgent  string = "chrome"
	FirefoxAgent string = "firefox"
	SafariAgent  string = "safari"
	UnknownAgent string = "unknown"
)

// ParseAgent parse browser long agent to short name
func ParseAgent(userAgent string) string {
	userAgent = strings.ToLower(userAgent)
	if strings.Contains(userAgent, "firefox/") {
		return FirefoxAgent
	}
	if strings.Contains(userAgent, "chrome/") {
		return ChromeAgent
	}
	return UnknownAgent
}
