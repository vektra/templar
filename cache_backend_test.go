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

func TestMemoryCacheBackend(t *testing.T) {
	n := neko.Start(t)

	var (
		cache *MemoryCacheBackend
	)

	n.Setup(func() {
		cache = NewMemoryCacheBackend(30 * time.Second)
	})

	n.It("can store and retrieve responses", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		upstream := &http.Response{
			Request:    req,
			StatusCode: 304,
			Status:     "304 Too Funky",
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

	n.Meow()
}
