package embedded_alias

import "go.uber.org/mock/mockgen/internal/tests/embedded_alias/other"

//go:generate mockgen -source=input.go -destination=mock_test.go -package=embedded_alias

type Example interface {
	LocalAlias
	other.Alias
}
