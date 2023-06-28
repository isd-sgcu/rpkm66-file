package gcs

import (
	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock
}

func (c *ClientMock) Upload(file []byte, _ string) error {
	args := c.Called(file)

	return args.Error(0)
}

func (c *ClientMock) GetSignedUrl(_ string) (string, error) {
	args := c.Called()

	return args.String(0), args.Error(1)
}
