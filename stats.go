package templar

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type DebugStats struct{}

func (d *DebugStats) Emit(req *http.Request, dur time.Duration) error {
	fmt.Printf("[%s] %s => %s\n", req.Method, req.URL, dur)
	return nil
}

type StatsdOutput struct {
	client StatsdClient
}

func NewStatsdOutput(client StatsdClient) *StatsdOutput {
	return &StatsdOutput{client}
}

func (s *StatsdOutput) url(req *http.Request) string {
	return req.Host + strings.Replace(req.URL.Path, "/", "-", -1)
}

func (s *StatsdOutput) StartRequest(req *http.Request) {
	s.client.Incr("templar.request.method."+req.Method, 1)
	s.client.Incr("templar.request.host."+req.Host, 1)
	s.client.Incr("templar.request.url."+s.url(req), 1)
	s.client.GaugeDelta("templar.requests.active", 1)
}

func (s *StatsdOutput) Emit(req *http.Request, delta time.Duration) {
	s.client.GaugeDelta("templar.requests.active", -1)
	s.client.PrecisionTiming("templar.request.url."+s.url(req), delta)
}
