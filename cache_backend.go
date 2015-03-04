package templar

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/vektra/templar/cache"
)

type MemoryCacheBackend struct {
	im cache.InMemoryCache
}

func NewMemoryCacheBackend(expire time.Duration) *MemoryCacheBackend {
	return &MemoryCacheBackend{
		im: cache.NewInMemoryCache(expire),
	}
}

type cachedRequest struct {
	body []byte
	resp *http.Response
}

func (m *MemoryCacheBackend) Set(req *http.Request, resp *http.Response) error {
	cr := &cachedRequest{}

	if resp.Body != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		cr.body = body
		resp.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	saved := &http.Response{}
	*saved = *resp

	cr.resp = saved

	m.im.Add(req.URL.String(), cr, 0)

	return nil
}

func (m *MemoryCacheBackend) Get(req *http.Request) (*http.Response, bool) {
	var cr *cachedRequest

	err := m.im.Get(req.URL.String(), &cr)

	if err != nil {
		return nil, false
	}

	saved := &http.Response{}

	*saved = *cr.resp

	saved.Body = ioutil.NopCloser(bytes.NewReader(cr.body))

	return saved, true
}
