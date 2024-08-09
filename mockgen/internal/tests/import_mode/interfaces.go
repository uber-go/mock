package import_mode

// This package is used for unit testing of import_mode.go.
// All the entities described here are inspired by the video game Helldivers 2

//go:generate mockgen -typed -package=mock -destination=mock/interfaces.go . DemocracyFan,SuperEarthCitizen,Projectile,Gun,Shooter,HelldiverRifleShooter,HelldiverRocketMan,PotentialTraitor,AgitationCampaign

import (
	"go.uber.org/mock/mockgen/internal/tests/import_mode/guns"
	"go.uber.org/mock/mockgen/internal/tests/import_mode/projectiles"
	. "go.uber.org/mock/mockgen/internal/tests/import_mode/stratagems"
)

type DemocracyFan interface {
	ILoveDemocracy()
	YouWillNeverDestroyOurWayOfLife()
}

type SuperEarthCitizen DemocracyFan

type Projectile interface {
	Speed() int
	FlightRange() int
	Explosive() bool
}

type Gun[ProjectileType Projectile] interface {
	Shoot(times int) []ProjectileType
	Ammo() int
}

type Shooter[ProjectileType Projectile, GunType Gun[ProjectileType]] interface {
	Gun() GunType
	Shoot(times int, targets ...*Enemy) (bool, error)
	Reload() bool
}

type Helldiver interface {
	DemocracyFan
	StratagemCarrier
}

type HelldiverRifleShooter interface {
	Helldiver
	Shooter[projectiles.Bullet, *guns.Rifle]
}

type HelldiverRocketMan interface {
	Helldiver
	Shooter[projectiles.Missle, *guns.RocketLauncher]
}

type PotentialTraitor interface {
	StratagemCarrier
	Shooter[projectiles.Bullet, *guns.Rifle]
}

type AgitationCampaign interface {
	interface {
		BecomeAHero()
		BecomeALegend()
		BecomeAHelldiver()
	}
}

type Enemy struct {
	Name     string
	Fraction string
	Hp       int
}

type Counter interface {
	int
}
