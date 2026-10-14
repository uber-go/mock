package embedded_alias

type Getter[T any] interface {
	Get() T
}

type LocalAlias = Getter[int]
