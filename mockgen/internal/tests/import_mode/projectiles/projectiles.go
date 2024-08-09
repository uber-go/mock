package projectiles

type Bullet struct{}

func (b Bullet) Speed() int {
	return 600
}

func (b Bullet) FlightRange() int {
	return 500
}

func (b Bullet) Explosive() bool {
	return false
}

type Missle struct {
}

func (m Missle) Speed() int {
	return 200
}

func (m Missle) FlightRange() int {
	return 800
}

func (m Missle) Explosive() bool {
	return true
}
