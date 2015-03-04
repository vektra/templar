package templar

import "net/http"

type Categorizer struct{}

const CategoryHeader = "X-Templar-Category"

func NewCategorizer() *Categorizer {
	return &Categorizer{}
}

func (c *Categorizer) Stateless(req *http.Request) bool {
	explicit := req.Header.Get(CategoryHeader)

	switch explicit {
	case "stateful":
		return false
	case "stateless":
		return true
	default:
		if req.Method == "GET" {
			return true
		}
	}

	return false
}
