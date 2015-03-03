package templar

import "net/http"

type FallbackCacher struct {
	backend   CacheBackend
	tranpsort Transport
}

func NewFallbackCacher(backend CacheBackend, transport Transport) *FallbackCacher {
	return &FallbackCacher{backend, transport}
}

func (c *FallbackCacher) RoundTrip(req *http.Request) (*http.Response, error) {
	upstream, err := c.tranpsort.RoundTrip(req)
	if err != nil {
		if err == ErrTimeout {
			if upstream, ok := c.backend.Get(req); ok {
				return upstream, nil
			}
		}

		return nil, err
	}

	c.backend.Set(req, upstream)

	return upstream, nil
}

type EagerCacher struct {
	backend   CacheBackend
	tranpsort Transport
}

func NewEagerCacher(backend CacheBackend, transport Transport) *EagerCacher {
	return &EagerCacher{backend, transport}
}

func (c *EagerCacher) RoundTrip(req *http.Request) (*http.Response, error) {
	if upstream, ok := c.backend.Get(req); ok {
		return upstream, nil
	}

	upstream, err := c.tranpsort.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	c.backend.Set(req, upstream)

	return upstream, nil
}
