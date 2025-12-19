package empty_interface

//go:generate mockgen -package empty_interface -destination mock.go -source input.go

type Empty any // migrating interface{} -> any does not resolve to an interface type.
