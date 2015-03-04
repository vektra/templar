package templar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestUpstream(t *testing.T) {
	n := neko.Start(t)

	var mockTrans MockTransport

	n.CheckMock(&mockTrans.Mock)

	n.It("sends a request to the transport", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := newRecordingSender()

		timeout := NewUpstream(&mockTrans)

		upstream := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
		}

		mockTrans.On("RoundTrip", req).Return(upstream, nil)

		err = timeout.Forward(res, req)
		require.NoError(t, err)

		assert.Equal(t, 304, res.w.Code)
	})

	n.It("will timeout a request if requested", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Add("X-Templar-Timeout", "2s")

		res := newRecordingSender()

		timeout := NewUpstream(&slowTransport{10})

		err = timeout.Forward(res, req)
		require.NoError(t, err)

		assert.Equal(t, 504, res.w.Code)
	})

	n.Meow()
}
