package import_mode

import (
	"go.uber.org/mock/mockgen/internal/tests/generics"
)

var bar generics.Bar[int, int] = &MockBar[int, int]{}

var universe generics.Universe[int] = &MockUniverse[int]{}

var solarSystem generics.SolarSystem[int] = &MockSolarSystem[int]{}

var earth generics.Earth[int] = &MockEarth[int]{}

var water generics.Water[int, uint] = &MockWater[int, uint]{}

var externalConstraint generics.ExternalConstraint[int64, int] = &MockExternalConstraint[int64, int]{}

var embeddedIface generics.EmbeddingIface[int, float64] = &MockEmbeddingIface[int, float64]{}

var generator generics.Generator[int] = &MockGenerator[int]{}

var group generics.Group[generics.Generator[any]] = &MockGroup[generics.Generator[any]]{}
