package subpkg

import aliased_time_import "time"

// Embedded is embedded by an interface in the parent package. Its method
// references time.Time through an aliased import that lives only in this
// (alphabetically later) file. See uber-go/mock#254.
type Embedded interface {
	Foo(t aliased_time_import.Time)
}
