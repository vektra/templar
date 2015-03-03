package templar

import "github.com/stretchr/testify/mock"

import "net/http"

type MockCacheBackend struct {
	mock.Mock
}

func (m *MockCacheBackend) Set(req *http.Request, resp *http.Response) {
	m.Called(req, resp)
}
func (m *MockCacheBackend) Get(req *http.Request) (*http.Response, bool) {
	ret := m.Called(req)

	r0 := ret.Get(0).(*http.Response)
	r1 := ret.Get(1).(bool)

	return r0, r1
}
