package aliased_embedded_interface

import "go.uber.org/mock/mockgen/internal/tests/aliased_embedded_interface/subpkg"

type Embedder interface {
	subpkg.Embedded
}
