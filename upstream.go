package templar

import (
	"errors"
	"net/http"
	"time"
)

const TemplarPrefix = "X-Templar-"

type Upstream struct {
	transport Transport
	stats     Stats
}

func NewUpstream(client Transport, stats Stats) *Upstream {
	return &Upstream{client, stats}
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

func (t *Upstream) forward(res Responder, req *http.Request) error {
	upstream, err := t.transport.RoundTrip(req)
	if err != nil {
		return err
	}

	w := res.Send(upstream)

	CopyBody(w, upstream.Body)

	return err
}

var ErrTimeout = errors.New("request timed out")

func (t *Upstream) Forward(res Responder, req *http.Request) error {
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
		t.stats.RequestTimeout(req, dur)

		if fb, ok := t.transport.(Fallback); ok {
			upstream, err := fb.Fallback(req)
			if err != nil {
				return err
			}

			if upstream != nil {
				CopyBody(res.Send(upstream), upstream.Body)
				return nil
			}
		}

		uperr := &http.Response{
			Request:    req,
			StatusCode: 504,
			Header:     make(http.Header),
		}

		uperr.Header.Set("X-Templar-TimedOut", "true")
		res.Send(uperr)
	}

	return nil
}
