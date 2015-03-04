package templar

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestCache(t *testing.T) {
	n := neko.Start(t)

	var (
		cache *Cache
	)

	n.Setup(func() {
		cache = NewMemoryCache(30 * time.Second)
	})

	n.It("can store and retrieve responses", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		upstream := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
			Header:     make(http.Header),
		}

		cache.Set(req, upstream)

		out, ok := cache.Get(req)
		require.True(t, ok)

		assert.Equal(t, upstream.StatusCode, out.StatusCode)
		assert.Equal(t, upstream.Header, out.Header)
	})

	n.It("makes the response body readable", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		funky := "waaay too funky"

		upstream := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
			Body:       ioutil.NopCloser(strings.NewReader(funky)),
		}

		cache.Set(req, upstream)

		_, err = ioutil.ReadAll(upstream.Body)
		require.NoError(t, err)

		out, ok := cache.Get(req)
		require.True(t, ok)

		bytes, err := ioutil.ReadAll(out.Body)
		require.NoError(t, err)

		assert.Equal(t, funky, string(bytes))
	})

	n.It("honors cache time requested in header", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		upstream := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
		}

		req.Header.Set(CacheTimeHeader, "1s")

		cache.Set(req, upstream)

		time.Sleep(1 * time.Second)

		_, ok := cache.Get(req)
		require.False(t, ok)
	})

	n.Meow()
}
