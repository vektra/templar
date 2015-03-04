package templar

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestUpstream(t *testing.T) {
	n := neko.Start(t)

	var (
		mockTrans MockTransport
		stats     MockStats
	)

	n.CheckMock(&mockTrans.Mock)
	n.CheckMock(&stats.Mock)

	n.It("sends a request to the transport", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := newRecordingSender()

		upstream := NewUpstream(&mockTrans, &stats)

		resp := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
		}

		mockTrans.On("RoundTrip", req).Return(resp, nil)

		err = upstream.Forward(res, req)
		require.NoError(t, err)

		assert.Equal(t, 304, res.w.Code)
	})

	n.It("does not send templar headers", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Set(CategoryHeader, "funky")

		trans := &HTTPTransport{&mockTrans}

		resp := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
		}

		exp, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		mockTrans.On("RoundTrip", exp).Return(resp, nil)

		_, err = trans.RoundTrip(req)
		require.NoError(t, err)
	})

	n.It("will timeout a request if requested", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Add("X-Templar-Timeout", "2s")

		res := newRecordingSender()

		upstream := NewUpstream(&slowTransport{10}, &stats)

		stats.On("RequestTimeout", req, 2*time.Second).Return(nil)

		err = upstream.Forward(res, req)
		require.NoError(t, err)

		assert.Equal(t, 504, res.w.Code)
	})

	n.It("will invoke a transports fallback on timeout", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Add("X-Templar-Timeout", "2s")

		res := newRecordingSender()

		upstream := NewUpstream(&slowTransportFallback{seconds: 10, fallback: true}, &stats)

		stats.On("RequestTimeout", req, 2*time.Second).Return(nil)

		err = upstream.Forward(res, req)
		require.NoError(t, err)

		assert.Equal(t, 201, res.w.Code)
	})

	n.It("handles the fallback indicating there is no fallback", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Add("X-Templar-Timeout", "2s")

		res := newRecordingSender()

		upstream := NewUpstream(&slowTransportFallback{seconds: 10, fallback: false}, &stats)

		stats.On("RequestTimeout", req, 2*time.Second).Return(nil)

		err = upstream.Forward(res, req)
		require.NoError(t, err)

		assert.Equal(t, 504, res.w.Code)
	})

	n.Meow()
}
