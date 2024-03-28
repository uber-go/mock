package overlap

import (
	"errors"
	"testing"

	gomock "go.uber.org/mock/gomock"
)

// TestValidInterface assesses whether or not the generated mock is valid
func TestValidInterface(t *testing.T) {
	ctrl := gomock.NewController(t)

	s := NewMockReadWriteCloser(ctrl)
	s.EXPECT().Close().Return(errors.New("test"))

	s.Close()
}
