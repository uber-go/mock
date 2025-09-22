package main

import (
	"iter"

	"go.uber.org/mock/mockgen/internal/tests/signature"
	"go.uber.org/mock/mockgen/internal/tests/signature/package_mode"
	"go.uber.org/mock/mockgen/internal/tests/signature/source_mode"
)

func UseSome(fnc signature.Some) {}

func UseSome2(fnc signature.Some2) {}

func UseGeneric[T any](fnc signature.Generic[T]) {}

func UseWithMethod(fnc signature.WithMethod) {}

func UseSeq[T any](fnc iter.Seq[T]) {}

func packageModeChecks() {
	UseSome((&package_mode.MockSome{}).Execute)
	UseSome((&package_mode.MockSome2{}).Execute)
	UseSome((&package_mode.MockAlias{}).Execute)
	UseSome2((&package_mode.MockSome{}).Execute)
	UseSome2((&package_mode.MockSome2{}).Execute)
	UseSome2((&package_mode.MockAlias{}).Execute)

	UseGeneric((&package_mode.MockGeneric[int]{}).Execute)
	UseGeneric[int]((&package_mode.MockIntGeneric{}).Execute)

	UseWithMethod((&package_mode.MockWithMethod{}).Execute)

	UseSeq((&package_mode.MockIntIter{}).Execute)
	UseSeq((&package_mode.MockSeq[bool]{}).Execute)
}

func sourceModeChecks() {
	// Source Mode currently does not support aliases and derived types.
	// Please uncomment when the opportunity arises.
	//UseSome((&source_mode.MockSome2{}).Execute)
	//UseSome((&source_mode.MockAlias{}).Execute)
	//UseSome2((&source_mode.MockSome2{}).Execute)
	//UseSome2((&source_mode.MockAlias{}).Execute)
	//UseGeneric[int]((&source_mode.MockIntGeneric{}).Execute)

	UseSome((&source_mode.MockSome{}).Execute)
	UseSome2((&source_mode.MockSome{}).Execute)
	UseGeneric((&source_mode.MockGeneric[int]{}).Execute)
	UseWithMethod((&source_mode.MockWithMethod{}).Execute)
}

// We check in compile-time that mocks are generated correctly
func main() {
	packageModeChecks()
	sourceModeChecks()
}
