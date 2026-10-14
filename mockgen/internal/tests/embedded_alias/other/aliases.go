package other

type Mapper[K, V any] interface {
	Map(K) V
}

type Alias = Mapper[string, bool]
