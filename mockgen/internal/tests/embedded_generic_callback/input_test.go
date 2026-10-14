package mypackage

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestEmbeddedGenericCallback(t *testing.T) {
	m := NewMockMyInterface(gomock.NewController(t))
	m.EXPECT().OnEvent(gomock.Any()).Do(func(callback Callback[MyStruct]) {
		callback(MyStruct{})
	})

	// The generated method must accept Callback[MyStruct] from this package.
	var target MyInterface = m
	called := false
	target.OnEvent(func(MyStruct) {
		called = true
	})
	if !called {
		t.Fatal("OnEvent did not invoke the callback")
	}
}
