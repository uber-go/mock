package any_type

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestAnyService(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockAnyService(ctrl)

	mock.EXPECT().Do("hello").Return("world")
	if res := mock.Do("hello"); res != "world" {
		t.Fatalf("unexpected result: %v", res)
	}

	mock.EXPECT().DoVariadic(1, 2, 3).Return([]any{4, 5, 6})
	items := mock.DoVariadic(1, 2, 3)
	if len(items) != 3 || items[0] != 4 {
		t.Fatalf("unexpected items: %v", items)
	}

	chIn := make(<-chan any)
	chOut := make(chan<- any)
	mock.EXPECT().DoChan(chIn).Return(chOut)
	if out := mock.DoChan(chIn); out != chOut {
		t.Fatalf("unexpected chan result: %v", out)
	}

	mIn := map[string]any{"key": "value"}
	mOut := map[string]any{"res": 123}
	mock.EXPECT().DoMap(mIn).Return(mOut)
	if out := mock.DoMap(mIn); out["res"] != 123 {
		t.Fatalf("unexpected map result: %v", out)
	}
}

func TestEmptyAny(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEmptyAny(ctrl)
	if mock == nil {
		t.Fatal("expected non-nil mock")
	}
}

func TestEmptyAnyAlias(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEmptyAnyAlias(ctrl)
	if mock == nil {
		t.Fatal("expected non-nil mock")
	}
}
