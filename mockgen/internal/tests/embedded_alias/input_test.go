package embedded_alias

import (
	"testing"

	"go.uber.org/mock/gomock"
)

var _ Example = (*MockExample)(nil)

func TestEmbeddedAliases(t *testing.T) {
	m := NewMockExample(gomock.NewController(t))
	m.EXPECT().Get().Return(42)
	m.EXPECT().Map("key").Return(true)

	if got := m.Get(); got != 42 {
		t.Errorf("Get() = %d, want 42", got)
	}
	if !m.Map("key") {
		t.Error("Map() = false, want true")
	}
}
