package stratagems

import (
	"time"
)

type Stratagem interface {
	Call() error
	Timeout() time.Duration
}

type StratagemCarrier interface {
	AvailableStratagems() []Stratagem
}
