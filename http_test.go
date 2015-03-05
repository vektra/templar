package templar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestHTTP(t *testing.T) {
	n := neko.Start(t)

	var (
		mockTrans MockTransport
	)

	n.CheckMock(&mockTrans.Mock)

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

	n.It("upgrades to https on request", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Set(UpgradeHeader, "https")

		trans := &HTTPTransport{&mockTrans}

		resp := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
		}

		exp, err := http.NewRequest("GET", "https://google.com/foo/bar", nil)
		require.NoError(t, err)

		mockTrans.On("RoundTrip", exp).Return(resp, nil)

		_, err = trans.RoundTrip(req)
		require.NoError(t, err)
	})

	n.Meow()
}
