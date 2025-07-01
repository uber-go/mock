package package_mode

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/mock/mockgen/internal/tests/generics"
)

func TestMockEmbeddingIface_One(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockEmbeddingIface[int, float64](ctrl)
	m.EXPECT().One("foo").Return("bar")
	if v := m.One("foo"); v != "bar" {
		t.Errorf("One() = %v, want %v", v, "bar")
	}
}

func TestMockUniverse_Water(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockUniverse[int](ctrl)
	m.EXPECT().Water(1024)
	m.Water(1024)
}

func TestNewMockGroup_Join(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockGroup[generics.Generator[any]](ctrl)
	ctx := context.TODO()
	m.EXPECT().Join(ctx).Return(nil)
	if v := m.Join(ctx); v != nil {
		t.Errorf("Join() = %v, want %v", v, nil)
	}
}

func TestMockAliasOfInstantiatedGeneric(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockBarAliasIntString(ctrl)
	m.EXPECT().Three(10).Return("bar")

	alias := generics.Bar[int, string](m)
	if v := alias.Three(10); v != "bar" {
		t.Errorf("Three(10) = %v, want %v", v, "bar")
	}
}
