package gomock_test

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func TestEcho_NoOverride(t *testing.T) {
	ctrl := gomock.NewController(t, gomock.WithOverridableExpectations())
	mockIndex := NewMockFoo(ctrl)

	mockIndex.EXPECT().Bar(gomock.Any()).Return("foo")
	res := mockIndex.Bar("input")

	if res != "foo" {
		t.Fatalf("expected response to equal 'foo', got %s", res)
	}
}

func TestEcho_WithOverride_BaseCase(t *testing.T) {
	ctrl := gomock.NewController(t, gomock.WithOverridableExpectations())
	mockIndex := NewMockFoo(ctrl)

	// initial expectation set
	mockIndex.EXPECT().Bar(gomock.Any()).Return("foo")
	// override
	mockIndex.EXPECT().Bar(gomock.Any()).Return("bar")
	res := mockIndex.Bar("input")

	if res != "bar" {
		t.Fatalf("expected response to equal 'bar', got %s", res)
	}
}

func TestEcho_WithOverrideArgsAware_BaseCase(t *testing.T) {
	ctrl := gomock.NewController(t, gomock.WithOverridableExpectationsArgsAware())
	mockIndex := NewMockFoo(ctrl)

	// initial expectation set
	mockIndex.EXPECT().Bar("first").Return("first initial")
	// another expectation
	mockIndex.EXPECT().Bar("second").Return("second initial")
	// reset first expectation
	mockIndex.EXPECT().Bar("first").Return("first changed")

	res := mockIndex.Bar("first")

	if res != "first changed" {
		t.Fatalf("expected response to equal 'first changed', got %s", res)
	}

	res = mockIndex.Bar("second")
	if res != "second initial" {
		t.Fatalf("expected response to equal 'second initial', got %s", res)
	}
}

func TestEcho_WithOverrideArgsAware_OverrideEqualMatchersOnly(t *testing.T) {
	ctrl := gomock.NewController(t, gomock.WithOverridableExpectationsArgsAware())
	mockIndex := NewMockFoo(ctrl)

	// initial expectation set
	mockIndex.EXPECT().Bar("foo").Return("foo").Times(1)
	mockIndex.EXPECT().Bar(gomock.Any()).Return("bar").Times(1)

	res := mockIndex.Bar("foo")

	if res != "foo" {
		t.Fatalf("expected response to equal 'foo', got %s", res)
	}

	res = mockIndex.Bar("bar")
	if res != "bar" {
		t.Fatalf("expected response to equal 'bar', got %s", res)
	}
}
