//go:build go1.24

package generics

//go:generate mockgen --build_constraint go1.24 --destination=package_mode/mock_generics_alias_test.go --package=package_mode . BarAliasTString,BarAliasIntR,BarAliasIntBazR,BarAliasIntBazString

// BarAliasTString is a generic alias of a generic with one complete and one
// generic type arg.
type BarAliasTString[T any] = Bar[T, string]

// BarAliasIntR is a generic alias of a generic with a renamed type param.
type BarAliasIntR[Q any] = Bar[int, Q]

// BarAliasIntBazR is a generic alias of a generic alias with a generic type arg.
type BarAliasIntBazR[Q any] = Bar[int, Baz[Q]]

// BarAliasIntBazString is an alias of a generic alias.
type BarAliasIntBazString = BarAliasIntBazR[string]
