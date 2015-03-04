package templar

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestCategorize(t *testing.T) {
	n := neko.Start(t)

	n.It("indicates that a GET is stateless", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		cat := NewCategorizer()

		assert.True(t, cat.Stateless(req))
	})

	n.It("indicates that a POST is not stateless", func() {
		req, err := http.NewRequest("POST", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		cat := NewCategorizer()

		assert.False(t, cat.Stateless(req))
	})

	n.It("honors a header to override behavior on be stateful", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Add(CategoryHeader, "stateful")

		cat := NewCategorizer()

		assert.False(t, cat.Stateless(req))
	})

	n.It("honors a header to override behavior to be stateless", func() {
		req, err := http.NewRequest("POST", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		req.Header.Add(CategoryHeader, "stateless")

		cat := NewCategorizer()

		assert.True(t, cat.Stateless(req))
	})

	n.Meow()
}
