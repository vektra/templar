package templar

import (
	"io"
	"net/http"
	"time"
)

type Responder interface {
	Send(resp *http.Response) io.Writer
}

type Client interface {
	Forward(res Responder, req *http.Request) error
}

type Stats interface {
	StartRequest(req *http.Request)
	Emit(req *http.Request, dur time.Duration)
	RequestTimeout(req *http.Request, timeout time.Duration)
}

type Transport interface {
	RoundTrip(*http.Request) (*http.Response, error)
	CancelRequest(req *http.Request)
}

type Fallback interface {
	Fallback(*http.Request) (*http.Response, error)
}

type CacheBackend interface {
	Set(req *http.Request, resp *http.Response)
	Get(req *http.Request) (*http.Response, bool)
}

type Finisher interface {
	Finish()
}

type StatsdClient interface {
	Incr(name string, count int64) error
	GaugeDelta(name string, delta int64) error
	PrecisionTiming(name string, t time.Duration) error
}
