package templar

import "github.com/stretchr/testify/mock"

import "net/http"
import "time"

type MockStats struct {
	mock.Mock
}

func (m *MockStats) StartRequest(req *http.Request) {
	m.Called(req)
}
func (m *MockStats) Emit(req *http.Request, dur time.Duration) {
	m.Called(req, dur)
}
func (m *MockStats) RequestTimeout(req *http.Request, timeout time.Duration) {
	m.Called(req, timeout)
}
