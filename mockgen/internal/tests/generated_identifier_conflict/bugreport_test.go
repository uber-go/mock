package bugreport

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestExample_Method(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockExample(ctrl)
	m.EXPECT().Method(1, 2, 3, 4)

	m.Method(1, 2, 3, 4)
}

func TestExample_VarargMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := NewMockExample(ctrl)
	m.EXPECT().VarargMethod(1, 2, 3, 4, 6, 7)

	m.VarargMethod(1, 2, 3, 4, 6, 7)
}
