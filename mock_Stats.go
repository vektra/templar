package templar

import "github.com/stretchr/testify/mock"

import "net/http"
import "time"

type MockStats struct {
	mock.Mock
}

func (m *MockStats) Emit(req *http.Request, dur time.Duration) error {
	ret := m.Called(req, dur)

	r0 := ret.Error(0)

	return r0
}
