package cache

import (
	"bytes"
	"errors"
	"github.com/golang/groupcache"
	"io/ioutil"
	"net/http"
	"time"
)

type GroupCacheCache struct {
	g groupcache.Group
	t http.Transport
}

type cachedRequest struct {
	body    []byte
	status  int
	headers http.Header
}

func NewGroupCacheCache(urls string, defaultExpiration time.Duration, transport http.Transport) *GroupCacheCache {
	peers := groupcache.NewHTTPPool("localhost:8080")
	peers.Set("localhost:8081")
	getter := func(context groupcache.Context, k string, destination groupcache.Sink) error {
		req, ok := context.(http.Request)
		if !ok {
			return errors.New("failed to cast groupcache context to an http request")
		}

		upstream, err := transport.RoundTrip(&req)

		destination.SetBytes()
		return nil
	}
	group := groupcache.NewGroup("templar", 250, groupcache.GetterFunc(getter))
	return &GroupCacheCache{*group}
}

func (c *GroupCacheCache) Set(req *http.Request, resp *http.Response) {
	// intentionally does nothing:
	// groupcache doesn't support sets - just reads
	// as such, we don't support sets, and gets go through a fallback
	// to an underlying http transport
}

func (c *GroupCacheCache) Get(req *http.Request) (*http.Response, bool) {
	var context groupcache.Context = req
	var data []byte
	err := c.g.Get(context, req.URL.Path, groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		return nil, true
	} else {
		cr := &cachedRequest{}
		Deserialize(data, cr)
		resp := &http.Response{
			StatusCode: cr.status,
			Header:     make(http.Header),
		}
		for k, v := range cr.headers {
			resp.Header[k] = v
		}

		resp.Body = ioutil.NopCloser(bytes.NewReader(cr.body))

		return resp, false
	}
}
