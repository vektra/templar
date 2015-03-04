package templar

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestProxy(t *testing.T) {
	n := neko.Start(t)

	var (
		client MockClient
		stats  MockStats
		proxy  *Proxy
	)

	n.CheckMock(&client.Mock)
	n.CheckMock(&stats.Mock)

	n.Setup(func() {
		proxy = NewProxy(&client, &stats)
	})

	n.It("sends the request on to the target", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := httptest.NewRecorder()

		stats.On("StartRequest", req).Return(nil)
		client.On("Forward", mock.Anything, req).Return(nil)
		stats.On("Emit", req, mock.Anything).Return(nil)

		proxy.ServeHTTP(res, req)
	})

	n.Meow()
}
