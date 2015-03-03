package templar

import (
	"net/http"
	"time"
)

type Client interface {
	Forward(res http.ResponseWriter, req *http.Request) error
}

type Stats interface {
	Emit(res http.ResponseWriter, req *http.Request, dur time.Duration) error
}
