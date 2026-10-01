package any_type

//go:generate mockgen -package any_type -destination mock.go -source input.go

type EmptyAny any

type EmptyAnyAlias = any

type AnyService interface {
	any
	Do(x any) any
	DoChan(ch <-chan any) chan<- any
	DoMap(m map[string]any) map[string]any
	DoVariadic(items ...any) []any
}
