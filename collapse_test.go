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

func TestCollapse(t *testing.T) {
	n := neko.Start(t)

	var (
		collapse *Collapser
		client   MockClient
	)

	n.CheckMock(&client.Mock)

	n.Setup(func() {
		collapse = NewCollapser(&client)
	})

	n.It("sends a request on to the downstream client", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := httptest.NewRecorder()

		client.On("Forward", res, req).Return(nil)

		collapse.Forward(res, req)
	})

	n.It("reuses a ongoing request if possible", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := httptest.NewRecorder()

		collapse.running[req.URL.String()] = &RunningRequest{Request: req}

		c := make(chan struct{})

		go func() {
			collapse.Forward(res, req)
			c <- struct{}{}
		}()

		// simulate waiting on the upstream
		time.Sleep(10 * time.Millisecond)

		upres := &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     http.Header{"X-Templar-Check": []string{"ok"}},
			Body:       nil,
			Request:    req,
		}

		collapse.finish(req, upres)

		select {
		case <-time.NewTimer(1 * time.Second).C:
			t.Fail()
		case <-c:
			// all good
		}

		assert.Equal(t, res.Code, 200)
		assert.Equal(t, res.Body.String(), "HTTP/1.1 200 OK\r\nX-Templar-Check: ok\r\nContent-Length: 0\r\n\r\n")
	})

	n.Meow()
}
