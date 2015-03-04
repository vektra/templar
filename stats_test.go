package templar

import (
	"net/http"
	"testing"
	"time"

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

		client.On("Incr", "templar.request.method.GET", 1).Return(nil)
		client.On("Incr", "templar.request.host.google.com", 1).Return(nil)
		client.On("Incr", "templar.request.url.google.com-foo-bar", 1).Return(nil)
		client.On("GaugeDelta", "templar.requests.active", 1).Return(nil)

		output.StartRequest(req)
	})

	n.It("emits stats about a request on end", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		t := 1 * time.Second

		client.On("GaugeDelta", "templar.requests.active", -1).Return(nil)
		client.On("PrecisionTiming",
			"templar.request.url.google.com-foo-bar", t).Return(nil)

		output.Emit(req, t)
	})

	n.Meow()
}
