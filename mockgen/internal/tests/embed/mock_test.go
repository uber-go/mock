// This test is for when mock is same package as the source.
package embed_test

import (
	reflect "reflect"
	"testing"

	"go.uber.org/mock/gomock"
	"go.uber.org/mock/mockgen/internal/tests/embed"
)

func TestEmbed(t *testing.T) {
	hoge := embed.NewMockHoge(gomock.NewController(t))
	et := reflect.TypeOf((*embed.Hoge)(nil)).Elem()
	ht := reflect.TypeOf(hoge)
	if !ht.Implements(et) {
		t.Errorf("source interface has been not implemented")
	}
}
