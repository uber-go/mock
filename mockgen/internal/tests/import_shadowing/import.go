package import_shadowing

import (
	"go.uber.org/mock/mockgen/internal/tests/import_shadowing/foo"
)

//go:generate mockgen -destination mock.go -source import.go -package import_shadowing . Bar

type Bar interface {
	Hoge(foo string) foo.MyInt
}
