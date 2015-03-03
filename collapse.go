package templar

import (
	"net/http"
	"sync"
)

type RunningRequest struct {
	Request *http.Request

	others []chan *http.Response
}

type Collapser struct {
	client Client

	lock    sync.Mutex
	running map[string]*RunningRequest
}

func NewCollapser(client Client) *Collapser {
	return &Collapser{
		client:  client,
		running: make(map[string]*RunningRequest),
	}
}

func (c *Collapser) Forward(res http.ResponseWriter, req *http.Request) error {
	c.lock.Lock()

	if running, ok := c.running[req.URL.String()]; ok {
		myself := make(chan *http.Response)
		running.others = append(running.others, myself)

		c.lock.Unlock()

		upstream := <-myself

		upstream.Write(res)

		return nil
	} else {
		c.lock.Unlock()
	}

	return c.client.Forward(res, req)
}

func (c *Collapser) finish(req *http.Request, res *http.Response) {
	c.lock.Lock()

	key := req.URL.String()

	running, ok := c.running[key]
	if ok {
		delete(c.running, key)
	}

	c.lock.Unlock()

	if ok {
		running.finish(res)
	}
}

func (r *RunningRequest) finish(res *http.Response) {
	for _, c := range r.others {
		c <- res
	}
}
