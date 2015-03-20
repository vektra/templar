package templar

import "github.com/stretchr/testify/mock"
import "github.com/amir/raidman"

type MockRiemannClient struct {
	mock.Mock
}

func (r *MockRiemannClient) Send(e *raidman.Event) error {
	r.Called(e)
	return nil
}
