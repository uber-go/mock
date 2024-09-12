package private

import (
	"reflect"
	"testing"

	"go.uber.org/mock/gomock"
)

func Test_t(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := newMockHelloWorld(ctrl)

	mock.EXPECT().Hi().Return("Hi")

	actual := mock.Hi()
	if !reflect.DeepEqual(actual, "Hi") {
		t.Errorf("got %v, want %v", actual, "Hi")
	}
}
