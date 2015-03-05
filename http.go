package templar

import (
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const UpgradeHeader = "X-Templar-Upgrade"

func NewHTTPTransport() Transport {
	return &HTTPTransport{
		&http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

type HTTPTransport struct {
	h Transport
}

func (h *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	out := &http.Request{}
	*out = *req

	out.RequestURI = ""

	out.Header = make(http.Header)

	for k, v := range req.Header {
		if strings.HasPrefix(k, TemplarPrefix) {
			if k == UpgradeHeader && v[0] == "https" {
				u := &url.URL{}
				*u = *out.URL

				u.Scheme = "https"

				out.URL = u
			}

			continue
		}

		out.Header[k] = v
	}

	return h.h.RoundTrip(out)
}

func (h *HTTPTransport) CancelRequest(req *http.Request) {
	h.h.CancelRequest(req)
}

func CopyResponse(res http.ResponseWriter, upstream *http.Response) {
	for k, v := range upstream.Header {
		res.Header()[k] = v
	}

	res.WriteHeader(upstream.StatusCode)
	if upstream.Body != nil {
		io.Copy(res, upstream.Body)
		upstream.Body.Close()
	}
}

func CopyBody(dst io.Writer, src io.Reader) {
	if src != nil {
		io.Copy(dst, src)
	}

	if fin, ok := dst.(Finisher); ok {
		fin.Finish()
	}
}
