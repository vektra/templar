package templar

import (
	"net/http"
	"time"
)

type Proxy struct {
	client Client
	stats  Stats
}

func NewProxy(cl Client, stats Stats) *Proxy {
	return &Proxy{cl, stats}
}

func (p *Proxy) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	start := time.Now()

	p.client.Forward(res, req)

	p.stats.Emit(res, req, time.Since(start))
}
