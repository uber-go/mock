//go:build go1.24

package package_mode

import (
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/mock/mockgen/internal/tests/generics"
)

func TestMockGenericAliasOfGeneric(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockBarAliasIntR[string](ctrl)
	m.EXPECT().Three(10).Return("bar")

	alias := generics.Bar[int, string](m)
	if v := alias.Three(10); v != "bar" {
		t.Errorf("Three(10) = %v, want %v", v, "bar")
	}
}

func TestMockGenericAliasOfGenericWithRenamedTypeParam(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockBarAliasTString[int](ctrl)
	m.EXPECT().Three(10).Return("bar")

	alias := generics.Bar[int, string](m)
	if v := alias.Three(10); v != "bar" {
		t.Errorf("Three(10) = %v, want %v", v, "bar")
	}
}

func TestMockGenericAliasOfGenericWithGenericTypeArg(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockBarAliasIntBazR[string](ctrl)
	m.EXPECT().Three(10).Return(generics.Baz[string]{V: "bar"})

	alias := generics.Bar[int, generics.Baz[string]](m)
	if v := alias.Three(10); v != (generics.Baz[string]{V: "bar"}) {
		t.Errorf("Three(10) = %v, want %v", v, "bar")
	}
}

func TestMockAliasOfInstatiatedGenericAlias(t *testing.T) {
	ctrl := gomock.NewController(t)

	m := NewMockBarAliasIntBazString(ctrl)
	m.EXPECT().Three(10).Return(generics.Baz[string]{V: "bar"})

	alias := generics.Bar[int, generics.Baz[string]](m)
	if v := alias.Three(10); v != (generics.Baz[string]{V: "bar"}) {
		t.Errorf("Three(10) = %v, want %v", v, "bar")
	}
}
