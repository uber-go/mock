package mock

import (
	"go.uber.org/mock/mockgen/internal/tests/alias"
)

// This checks for type-checking equivalent of mock types.
// If something does not resolve, the tests will not compile.

var (
	_ alias.Fooer          = &MockFooer{}
	_ alias.FooerAlias     = &MockFooer{}
	_ alias.Fooer          = &MockFooerAlias{}
	_ alias.FooerAlias     = &MockFooerAlias{}
	_ alias.Barer          = &MockBarer{}
	_ alias.BarerAlias     = &MockBarer{}
	_ alias.Barer          = &MockBarerAlias{}
	_ alias.BarerAlias     = &MockBarerAlias{}
	_ alias.Bazer          = &MockBazer{}
	_ alias.QuxerConsumer  = &MockQuxerConsumer{}
	_ alias.QuuxerConsumer = &MockQuuxerConsumer{}
)
