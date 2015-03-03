package templar

import (
	"errors"
	"net/http"
	"time"
)

type Upstream struct {
	transport Transport
}

func NewUpstream(client Transport) *Upstream {
	return &Upstream{client}
}

const CTimeoutHeader = "X-Templar-Timeout"

func (t *Upstream) extractTimeout(req *http.Request) (time.Duration, bool) {
	header := req.Header.Get(CTimeoutHeader)
	if header == "" {
		return 0, false
	}

	dur, err := time.ParseDuration(header)
	if err != nil {
		return 0, false
	}

	return dur, true
}

func (t *Upstream) forward(res http.ResponseWriter, req *http.Request) error {
	upstream, err := t.transport.RoundTrip(req)
	if err != nil {
		return err
	}

	for k, v := range upstream.Header {
		res.Header()[k] = v
	}

	res.WriteHeader(upstream.StatusCode)
	return upstream.Write(res)
}

var ErrTimeout = errors.New("request timed out")

func (t *Upstream) Forward(res http.ResponseWriter, req *http.Request) error {
	dur, ok := t.extractTimeout(req)

	if !ok {
		return t.forward(res, req)
	}

	fin := make(chan error)

	go func() {
		fin <- t.forward(res, req)
	}()

	time.AfterFunc(dur, func() {
		t.transport.CancelRequest(req)
		fin <- ErrTimeout
	})

	err := <-fin

	if err == ErrTimeout {
		res.WriteHeader(504)
		res.Header().Set("X-Templar-TimedOut", "true")
	}

	return nil
}
