package cache

import (
	"bytes"
	"errors"
	"github.com/golang/groupcache"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Transport interface {
	RoundTrip(*http.Request) (*http.Response, error)
}

type GroupCacheCache struct {
	g groupcache.Group
	t Transport
}

type cachedResponse struct {
	Body    []byte
	Status  int
	Headers http.Header
}

func NewGroupCacheCache(thisPeerURL string, otherPeersURLs string, defaultExpiration time.Duration, transport Transport) *GroupCacheCache {
	otherPeers := strings.Split(otherPeersURLs, ",")
	pool := groupcache.NewHTTPPool(thisPeerURL)
	pool.Set(otherPeers...)
	getter := func(context groupcache.Context, k string, destination groupcache.Sink) error {
		req, ok := context.(*http.Request)
		if !ok {
			return errors.New("failed to cast groupcache context to an http request")
		}

		upstream, err := transport.RoundTrip(req)
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(upstream.Body)
		if err != nil {
			return err
		}
		toCache := &cachedResponse{
			Body:    body,
			Status:  upstream.StatusCode,
			Headers: upstream.Header,
		}

		b, err := Serialize(toCache)
		if err != nil {
			return err
		}
		destination.SetBytes(b)
		return nil
	}
	group := groupcache.NewGroup("templar", 64<<20, groupcache.GetterFunc(getter))
	return &GroupCacheCache{*group, transport}
}

func (c *GroupCacheCache) Set(req *http.Request, resp *http.Response) {
	// intentionally does nothing:
	// groupcache doesn't support sets - just reads
	// as such, we don't support sets, and gets go through a fallback
	// to an underlying http transport
}

func (c *GroupCacheCache) Get(req *http.Request) (*http.Response, bool) {
	var data []byte
	err := c.g.Get(req, req.URL.Path, groupcache.AllocatingByteSliceSink(&data))
	if err != nil {
		return nil, false
	} else {
		cr := &cachedResponse{}
		Deserialize(data, cr)
		resp := &http.Response{
			StatusCode: cr.Status,
			Header:     make(http.Header),
		}
		for k, v := range cr.Headers {
			resp.Header[k] = v
		}

		resp.Body = ioutil.NopCloser(bytes.NewReader(cr.Body))

		return resp, true
	}
}
