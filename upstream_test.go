package templar

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

type slowTransport struct{}

func (s slowTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	time.Sleep(10 * time.Second)
	return nil, nil
}

func (s slowTransport) CancelRequest(req *http.Request) {
}

func TestUpstreamTimeout(t *testing.T) {
	n := neko.Start(t)

	n.It("will timeout a request if requested", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Add("X-Templar-Timeout", "2s")

		res := httptest.NewRecorder()

		timeout := NewUpstream(slowTransport{})

		err = timeout.Forward(res, req)
		require.NoError(t, err)

		assert.Equal(t, 504, res.Code)
	})

	n.Meow()
}
