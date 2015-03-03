package templar

import "github.com/stretchr/testify/mock"

import "net/http"

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Forward(res http.ResponseWriter, req *http.Request) error {
	ret := m.Called(res, req)

	r0 := ret.Error(0)

	return r0
}
