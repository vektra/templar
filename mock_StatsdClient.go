package templar

import "github.com/stretchr/testify/mock"

import "time"

type MockStatsdClient struct {
	mock.Mock
}

func (m *MockStatsdClient) Incr(name string, count int64) error {
	ret := m.Called(name, count)

	r0 := ret.Error(0)

	return r0
}
func (m *MockStatsdClient) GaugeDelta(name string, delta int64) error {
	ret := m.Called(name, delta)

	r0 := ret.Error(0)

	return r0
}
func (m *MockStatsdClient) PrecisionTiming(name string, t time.Duration) error {
	ret := m.Called(name, t)

	r0 := ret.Error(0)

	return r0
}
