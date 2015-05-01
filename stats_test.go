package templar

import (
	"net/http"
	"testing"
	"time"

	"github.com/amir/raidman"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestStatsdOutput(t *testing.T) {
	n := neko.Start(t)

	var (
		output *StatsdOutput
		client MockStatsdClient
	)

	n.CheckMock(&client.Mock)

	n.Setup(func() {
		output = NewStatsdOutput(&client)
	})

	n.It("emits stats about a request on start", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		client.On("Incr", "templar.request.method.GET", int64(1)).Return(nil)
		client.On("Incr", "templar.request.host.google.com", int64(1)).Return(nil)
		client.On("Incr", "templar.request.url.google.com-foo-bar", int64(1)).Return(nil)
		client.On("GaugeDelta", "templar.requests.active", int64(1)).Return(nil)

		output.StartRequest(req)
	})

	n.It("emits stats about a request on end", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		t := 1 * time.Second

		client.On("GaugeDelta", "templar.requests.active", int64(-1)).Return(nil)
		client.On("PrecisionTiming",
			"templar.request.url.google.com-foo-bar", t).Return(nil)

		output.Emit(req, t)
	})

	n.It("emits stats when a request times out", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		t := 5 * time.Second

		client.On("Incr", "templar.timeout.host.google.com", int64(1)).Return(nil)
		client.On("Incr", "templar.timeout.url.google.com-foo-bar", int64(1)).Return(nil)

		output.RequestTimeout(req, t)

	})

	n.Meow()
}

func TestRiemannOutput(t *testing.T) {
	n := neko.Start(t)

	var (
		output *RiemannOutput
		client MockRiemannClient
	)

	n.CheckMock(&client.Mock)

	n.Setup(func() {
		output = NewRiemannOutput(&client)
	})

	n.It("emits stats about a request on start", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		attributes := make(map[string]string)
		attributes["method"] = "GET"
		attributes["host"] = "google.com"
		attributes["path"] = "/foo/bar"
		var event = &raidman.Event{
			State:      "ok",
			Service:    "templar request",
			Metric:     1,
			Attributes: attributes,
		}
		client.On("Send", event).Return(nil)

		output.StartRequest(req)
	})

	n.It("emits stats about a request on end", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		t := 1 * time.Second

		attributes := make(map[string]string)
		attributes["method"] = "GET"
		attributes["host"] = "google.com"
		attributes["path"] = "/foo/bar"
		var event = &raidman.Event{
			State:      "ok",
			Service:    "templar response",
			Metric:     1000.0,
			Attributes: attributes,
		}
		client.On("Send", event).Return(nil)

		output.Emit(req, t)
	})

	n.It("emits stats when a request times out", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		t := 5 * time.Second

		attributes := make(map[string]string)
		attributes["method"] = "GET"
		attributes["host"] = "google.com"
		attributes["path"] = "/foo/bar"
		var event = &raidman.Event{
			State:      "warning",
			Service:    "templar timeout",
			Metric:     5000.0,
			Attributes: attributes,
		}
		client.On("Send", event).Return(nil)

		output.RequestTimeout(req, t)
	})

	n.Meow()
}
