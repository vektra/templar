package templar

import "github.com/stretchr/testify/mock"

import "io"
import "net/http"

type MockResponder struct {
	mock.Mock
}

func (m *MockResponder) Send(resp *http.Response) io.Writer {
	ret := m.Called(resp)

	r0 := ret.Get(0).(io.Writer)

	return r0
}
