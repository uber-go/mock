package signature

import (
	"iter"
)

//go:generate mockgen -typed -destination=source_mode/mock.go -package=source_mode -source=signatures.go
//go:generate mockgen -typed -destination=package_mode/mock.go -package=package_mode . Some,Some2,Alias,Generic,IntGeneric,WithMethod,IntIter
//go:generate mockgen -typed -destination=package_mode/external_mock.go -package=package_mode iter Seq,Seq2

type Some func() string

type Some2 Some

type Alias = Some

type Generic[T any] func(T) string

type IntGeneric Generic[int]

type WithMethod func(string) string

func (f WithMethod) Method() {}

type IntIter iter.Seq[int]
