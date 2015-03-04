package templar

import "net/http"

type FallbackCacher struct {
	backend   CacheBackend
	tranpsort Transport

	categorizer *Categorizer
}

func NewFallbackCacher(backend CacheBackend, transport Transport, categorizer *Categorizer) *FallbackCacher {
	return &FallbackCacher{backend, transport, categorizer}
}

func (c *FallbackCacher) RoundTrip(req *http.Request) (*http.Response, error) {
	upstream, err := c.tranpsort.RoundTrip(req)

	if !c.categorizer.Stateless(req) {
		return upstream, err
	}

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

	categorizer *Categorizer
}

func NewEagerCacher(backend CacheBackend, transport Transport, categorizer *Categorizer) *EagerCacher {
	return &EagerCacher{backend, transport, categorizer}
}

func (c *EagerCacher) RoundTrip(req *http.Request) (*http.Response, error) {
	if !c.categorizer.Stateless(req) {
		upstream, err := c.tranpsort.RoundTrip(req)
		return upstream, err
	}

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
