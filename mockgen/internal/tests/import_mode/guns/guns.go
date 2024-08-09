package guns

import (
	"go.uber.org/mock/mockgen/internal/tests/import_mode/projectiles"
)

type Rifle struct {
	ammoSize int
}

func (r *Rifle) Shoot(times int) []projectiles.Bullet {
	currentAmmo := r.ammoSize
	r.ammoSize = min(0, currentAmmo-times)

	return make([]projectiles.Bullet, min(currentAmmo, times))
}

func (r *Rifle) Ammo() int {
	return r.ammoSize
}

type RocketLauncher struct {
	loaded bool
}

func (r *RocketLauncher) Shoot(times int) []projectiles.Missle {
	if r.loaded {
		r.loaded = false
		return make([]projectiles.Missle, 1)
	}

	return nil
}

func (r *RocketLauncher) Ammo() int {
	if r.loaded {
		return 1
	}

	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
