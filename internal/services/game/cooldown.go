package game

import "time"

const (
	maxAttemptsPerSpawn = 10 // allow 10 shots per spawn
	shootCooldown       = 5 * time.Second
)

// Shot state per player
type shotState struct {
	SpawnID       int64
	Attempts      int
	CooldownUntil time.Time
}

// canShoot allows up to maxAttemptsPerSpawn shots per spawn.
// Attempts reset instantly when a new pigeon spawns (spawnID changes).
// No time-based cooldown at all.
func (g *Game) canShoot(name string, spawnID int64) (bool, time.Duration) {
	g.shotMu.Lock()
	defer g.shotMu.Unlock()

	now := time.Now()

	st := g.lastShot[name]
	if st == nil {
		st = &shotState{}
		g.lastShot[name] = st
	}

	if st.SpawnID != spawnID {
		st.SpawnID = spawnID
		st.Attempts = 0
		st.CooldownUntil = time.Time{} // ✅ ล้าง cooldown
	}

	if !st.CooldownUntil.IsZero() && now.Before(st.CooldownUntil) {
		return false, time.Until(st.CooldownUntil)
	}

	// Allow 5 shots, then start cooldown but DO NOT require new spawn
	if st.Attempts >= maxAttemptsPerSpawn {
		st.Attempts = 0 // reset attempts after cooldown starts
		st.CooldownUntil = now.Add(shootCooldown)
		return false, shootCooldown
	}

	st.Attempts++
	return true, 0
}
