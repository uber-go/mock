package import_aliased

import (
	definition_alias "github.com/package/definition"
)

//go:generate mockgen -package import_aliased -destination source_mock.go -source=source.go -imports definition_alias=github.com/package/definition

type S interface {
	M(definition_alias.X)
}
