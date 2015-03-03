package templar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestFallbackCache(t *testing.T) {
	n := neko.Start(t)

	var (
		backend   MockCacheBackend
		transport MockTransport
		cache     *FallbackCacher
	)

	n.CheckMock(&backend.Mock)
	n.CheckMock(&transport.Mock)

	n.Setup(func() {
		cache = NewFallbackCacher(&backend, &transport)
	})

	n.It("caches a response that flows though it", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		upstream := &http.Response{
			Request: req,
		}

		transport.On("RoundTrip", req).Return(upstream, nil)
		backend.On("Set", req, upstream).Return(nil)

		_, err = cache.RoundTrip(req)
		require.NoError(t, err)
	})

	n.It("retrieves the value from the cache if there is a timeout", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		upstream := &http.Response{
			Request: req,
		}

		transport.On("RoundTrip", req).Return((*http.Response)(nil), ErrTimeout)
		backend.On("Get", req).Return(upstream, true)

		out, err := cache.RoundTrip(req)
		require.NoError(t, err)

		assert.Equal(t, upstream, out)
	})

	n.Meow()
}

func TestEagerCache(t *testing.T) {
	n := neko.Start(t)

	var (
		backend   MockCacheBackend
		transport MockTransport
		cache     *EagerCacher
	)

	n.CheckMock(&backend.Mock)
	n.CheckMock(&transport.Mock)

	n.Setup(func() {
		cache = NewEagerCacher(&backend, &transport)
	})

	n.It("caches a response that flows though it", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		upstream := &http.Response{
			Request: req,
		}

		backend.On("Get", req).Return((*http.Response)(nil), false)
		transport.On("RoundTrip", req).Return(upstream, nil)
		backend.On("Set", req, upstream).Return(nil)

		_, err = cache.RoundTrip(req)
		require.NoError(t, err)
	})

	n.It("retrieves the value from the cache if there is a timeout", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		upstream := &http.Response{
			Request: req,
		}

		backend.On("Get", req).Return(upstream, true)

		out, err := cache.RoundTrip(req)
		require.NoError(t, err)

		assert.Equal(t, upstream, out)
	})

	n.Meow()
}
