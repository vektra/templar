package templar

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

type recordingSender struct {
	w *httptest.ResponseRecorder
}

func newRecordingSender() *recordingSender {
	return &recordingSender{httptest.NewRecorder()}
}

func (r *recordingSender) Send(res *http.Response) io.Writer {
	for k, v := range res.Header {
		r.w.Header()[k] = v
	}

	r.w.WriteHeader(res.StatusCode)

	return r.w
}

type slowTransport struct {
	seconds time.Duration
}

func (st *slowTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response := &http.Response{
		Request:    req,
		StatusCode: 200,
		Body:       ioutil.NopCloser(strings.NewReader(fmt.Sprintf("now: %s", time.Now()))),
	}

	time.Sleep(st.seconds * time.Second)

	return response, nil
}

func (st *slowTransport) CancelRequest(req *http.Request) {}
