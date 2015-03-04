package templar

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/vektra/templar/cache"
)

type Cache struct {
	c cache.Cache
}

func NewMemoryCache(expire time.Duration) *Cache {
	return &Cache{
		c: cache.NewInMemoryCache(expire),
	}
}

func NewMemcacheCache(hostlist []string, expire time.Duration) *Cache {
	return &Cache{
		c: cache.NewMemcachedCache(hostlist, expire),
	}
}

func NewRedisCache(host string, password string, expire time.Duration) *Cache {
	return &Cache{
		c: cache.NewRedisCache(host, password, expire),
	}
}

type cachedRequest struct {
	body []byte
	resp *http.Response
}

func (m *Cache) Set(req *http.Request, resp *http.Response) error {
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

	m.c.Add(req.URL.String(), cr, 0)

	return nil
}

func (m *Cache) Get(req *http.Request) (*http.Response, bool) {
	var cr *cachedRequest

	err := m.c.Get(req.URL.String(), &cr)

	if err != nil {
		return nil, false
	}

	saved := &http.Response{}

	*saved = *cr.resp

	saved.Body = ioutil.NopCloser(bytes.NewReader(cr.body))

	return saved, true
}
