package game

import (
	"time"
)

const shootCooldown = 5 * time.Second

func (g *Game) canShoot(name string) (bool, time.Duration) {
	g.shotMu.Lock()
	defer g.shotMu.Unlock()

	now := time.Now()

	if last, ok := g.lastShot[name]; ok {
		if now.Sub(last) < shootCooldown {
			return false, shootCooldown - now.Sub(last)
		}
	}

	g.lastShot[name] = now
	return true, 0
}
