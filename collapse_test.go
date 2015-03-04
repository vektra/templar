package templar

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

type heldClient struct {
	begin chan struct{}
	wait  chan struct{}
}

func (h *heldClient) Finish() {
	h.wait <- struct{}{}
}

func (h *heldClient) Forward(res Responder, req *http.Request) error {
	h.begin <- struct{}{}
	<-h.wait

	return nil
}

func TestCollapse(t *testing.T) {
	n := neko.Start(t)

	var (
		collapse *Collapser
		client   MockClient
	)

	n.CheckMock(&client.Mock)

	n.Setup(func() {
		collapse = NewCollapser(&client, NewCategorizer())
	})

	n.It("sends a request on to the downstream client", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := newRecordingSender()

		client.On("Forward", mock.Anything, req).Return(nil)

		collapse.Forward(res, req)
	})

	n.It("registers itself as an ongoing request", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := newRecordingSender()

		held := &heldClient{
			make(chan struct{}),
			make(chan struct{}),
		}

		defer held.Finish()

		collapse := NewCollapser(held, NewCategorizer())

		go func() {
			collapse.Forward(res, req)
		}()

		<-held.begin

		_, ok := collapse.running[req.URL.String()]
		assert.True(t, ok)
	})

	n.It("reuses a ongoing request if possible", func() {
		req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		parentRes := newRecordingSender()
		childRes := newRecordingSender()

		rr := &RunningRequest{
			Request: req,
			others:  []Responder{parentRes},
			done:    make(chan struct{}),
		}

		collapse.running[req.URL.String()] = rr

		c := make(chan struct{})

		go func() {
			collapse.Forward(childRes, req)
			c <- struct{}{}
		}()

		// simulate waiting on the upstream
		time.Sleep(10 * time.Millisecond)

		upres := &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     http.Header{"X-Templar-Check": []string{"ok"}},
			Body:       nil,
			Request:    req,
		}

		io := collapse.finish(req, upres, rr)
		CopyBody(io, upres.Body)

		select {
		case <-time.NewTimer(1 * time.Second).C:
			t.Fatal()
		case <-c:
			// all good
		}

		assert.Equal(t, childRes.w.Code, 200)
		assert.Equal(t, "ok", childRes.w.Header().Get("X-Templar-Check"))
	})

	n.It("does not collapse stateful requests", func() {
		req, err := http.NewRequest("POST", "http://google.com/foo/bar", nil)
		require.NoError(t, err)

		res := newRecordingSender()

		client.On("Forward", mock.Anything, req).Return(nil)

		collapse.Forward(res, req)

		assert.Equal(t, len(collapse.running), 0)
	})

	n.It("can collapse multiple requests into one", func() {
		upstream := NewUpstream(&slowTransport{1})

		collapse := NewCollapser(upstream, NewCategorizer())

		var wg sync.WaitGroup

		var responses []*recordingSender

		for i := 0; i < 10; i++ {
			req, err := http.NewRequest("GET", "http://google.com/foo/bar", nil)
			require.NoError(t, err)

			res := newRecordingSender()

			responses = append(responses, res)

			wg.Add(1)
			go func(res Responder, req *http.Request) {
				defer wg.Done()

				collapse.Forward(res, req)
			}(res, req)
		}

		wg.Wait()

		first := responses[0].w.Body.String()

		assert.True(t, len(first) != 0)

		for _, resp := range responses {
			assert.Equal(t, first, resp.w.Body.String())
		}
	})

	n.Meow()
}
