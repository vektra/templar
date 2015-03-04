package templar

import (
	"fmt"
	"net/http"
	"time"
)

type DebugStats struct{}

func (d *DebugStats) Emit(res http.ResponseWriter, req *http.Request, dur time.Duration) error {
	fmt.Printf("[%s] %s => %s\n", req.Method, req.URL, dur)
	return nil
}
