package templar

import "net/http"

type FallbackCacher struct {
	backend   CacheBackend
	transport Transport

	categorizer *Categorizer
}

var _ = Transport(&FallbackCacher{})

func NewFallbackCacher(backend CacheBackend, transport Transport, categorizer *Categorizer) *FallbackCacher {
	return &FallbackCacher{backend, transport, categorizer}
}

const (
	CacheHeader       = "X-Templar-Cache"
	CacheTimeHeader   = "X-Templar-CacheFor"
	CacheCachedHeader = "X-Templar-Cached"
)

func (c *FallbackCacher) shouldCache(req *http.Request) bool {
	return c.categorizer.Stateless(req) && req.Header.Get(CacheHeader) == "fallback"
}

func (c *FallbackCacher) RoundTrip(req *http.Request) (*http.Response, error) {
	upstream, err := c.transport.RoundTrip(req)

	if !c.shouldCache(req) {
		return upstream, err
	}

	if err != nil {
		return nil, err
	}

	c.backend.Set(req, upstream)

	return upstream, nil
}

func (c *FallbackCacher) Fallback(req *http.Request) (*http.Response, error) {
	if upstream, ok := c.backend.Get(req); ok {
		if upstream != nil {
			upstream.Header.Add(CacheCachedHeader, "yes")
			return upstream, nil
		}
	}

	return nil, nil
}

func (c *FallbackCacher) CancelRequest(req *http.Request) {
	c.transport.CancelRequest(req)
}

type EagerCacher struct {
	backend   CacheBackend
	transport Transport

	categorizer *Categorizer
}

var _ = Transport(&EagerCacher{})

func NewEagerCacher(backend CacheBackend, transport Transport, categorizer *Categorizer) *EagerCacher {
	return &EagerCacher{backend, transport, categorizer}
}

func (c *EagerCacher) shouldCache(req *http.Request) bool {
	return c.categorizer.Stateless(req) && req.Header.Get(CacheHeader) == "eager"
}

func (c *EagerCacher) RoundTrip(req *http.Request) (*http.Response, error) {
	if !c.shouldCache(req) {
		upstream, err := c.transport.RoundTrip(req)
		return upstream, err
	}

	if upstream, ok := c.backend.Get(req); ok {
		upstream.Header.Add(CacheCachedHeader, "yes")

		return upstream, nil
	}

	upstream, err := c.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	c.backend.Set(req, upstream)

	return upstream, nil
}

func (c *EagerCacher) CancelRequest(req *http.Request) {
	c.transport.CancelRequest(req)
}
