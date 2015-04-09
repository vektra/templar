package cache

import (
	"bytes"
	"errors"
	"github.com/golang/groupcache"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	CacheTimeHeader = "X-Templar-CacheFor"
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

func calculateEpochedKey(req *http.Request, now time.Time) string {
	expires := FOREVER
	if reqExpire := req.Header.Get(CacheTimeHeader); reqExpire != "" {
		if dur, err := time.ParseDuration(reqExpire); err == nil {
			expires = dur
		}
	}
	if expires != FOREVER {
		return req.URL.String() + strconv.Itoa(int(now.Round(expires).Unix()))
	} else {
		return req.URL.String()
	}
}

func (c *GroupCacheCache) Get(req *http.Request) (*http.Response, bool) {
	var data []byte
	key := calculateEpochedKey(req, time.Now())
	err := c.g.Get(req, key, groupcache.AllocatingByteSliceSink(&data))
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
