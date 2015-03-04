package templar

import "net/http"

type Categorizer struct{}

func NewCategorizer() *Categorizer {
	return &Categorizer{}
}

func (c *Categorizer) Stateless(req *http.Request) bool {
	if req.Method == "GET" {
		return true
	}

	return false
}
