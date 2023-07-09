package source

import (
	"testing"

	"go.uber.org/mock/mockgen/internal/tests/generics"
)

func TestAssert(t *testing.T) {
	var x MockEmbeddingIface[int, float64]
	var _ generics.EmbeddingIface[int, float64] = &x
}
