package templar

import (
	"net/http"
	"time"
)

type Client interface {
	Forward(res http.ResponseWriter, req *http.Request) error
}

type Stats interface {
	Emit(res http.ResponseWriter, req *http.Request, dur time.Duration) error
}

type Transport interface {
	RoundTrip(*http.Request) (*http.Response, error)
	CancelRequest(req *http.Request)
}

type CacheBackend interface {
	Set(req *http.Request, resp *http.Response)
	Get(req *http.Request) (*http.Response, bool)
}
