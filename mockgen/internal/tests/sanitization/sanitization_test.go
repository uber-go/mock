package sanitization

import (
	"testing"

	"go.uber.org/mock/gomock"
	any0 "go.uber.org/mock/mockgen/internal/tests/sanitization/any"
	"go.uber.org/mock/mockgen/internal/tests/sanitization/mockout"
)

func TestSanitization(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := mockout.NewMockAnyMock(ctrl)
	m.EXPECT().Do(gomock.Any(), 1)
	m.Do(&any0.Any{}, 1)
}
