package templar

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func NewHTTPTransport() Transport {
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

func CopyResponse(res http.ResponseWriter, upstream *http.Response) {
	for k, v := range upstream.Header {
		res.Header()[k] = v
	}

	res.WriteHeader(upstream.StatusCode)
	if upstream.Body != nil {
		fmt.Printf("copy upstream... %#v\n", upstream.Body)
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
