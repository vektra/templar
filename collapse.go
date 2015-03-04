package templar

import (
	"io"
	"net/http"
	"sync"
)

type RunningRequest struct {
	Request *http.Request

	others []Responder
	done   chan struct{}
}

type Collapser struct {
	client      Client
	categorizer *Categorizer

	lock    sync.Mutex
	running map[string]*RunningRequest
}

func NewCollapser(client Client, categorizer *Categorizer) *Collapser {
	return &Collapser{
		client:      client,
		categorizer: categorizer,
		running:     make(map[string]*RunningRequest),
	}
}

type collapseResponder struct {
	collapser *Collapser
	request   *http.Request
	running   *RunningRequest
}

func (c *collapseResponder) Send(res *http.Response) io.Writer {
	return c.collapser.finish(c.request, res, c.running)
}

func (c *Collapser) finish(req *http.Request, res *http.Response, rr *RunningRequest) io.Writer {

	c.lock.Lock()

	key := req.URL.String()
	delete(c.running, key)

	c.lock.Unlock()

	cw := &collapsedWriter{running: rr}

	for _, c := range rr.others {
		cw.w = append(cw.w, c.Send(res))
	}

	return cw
}

type collapsedWriter struct {
	w       []io.Writer
	running *RunningRequest
}

func (cw *collapsedWriter) Write(p []byte) (n int, err error) {
	for _, w := range cw.w {
		n, err = w.Write(p)
		if err != nil {
			return
		}

		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}

	return len(p), nil
}

func (cw *collapsedWriter) Finish() {
	close(cw.running.done)
}

func (c *Collapser) Forward(res Responder, req *http.Request) error {
	if !c.categorizer.Stateless(req) {
		return c.client.Forward(res, req)
	}

	c.lock.Lock()

	key := req.URL.String()

	if running, ok := c.running[key]; ok {
		running.others = append(running.others, res)

		c.lock.Unlock()

		<-running.done

		return nil
	}

	rr := &RunningRequest{
		Request: req,
		others:  []Responder{res},
		done:    make(chan struct{}),
	}

	c.running[key] = rr
	c.lock.Unlock()

	return c.client.Forward(&collapseResponder{c, req, rr}, req)
}
