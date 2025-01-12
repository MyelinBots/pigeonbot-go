package player

// Player represents a game player
type Player struct {
	Name   string
	Points int
	Count  int
}

// StartingPoints is the default initial points for players
const StartingPoints = 0

// NewPlayer initializes a new Player instance
func NewPlayer(name string, startingPoints int, startingCount int) *Player {
	return &Player{
		Name:   name,
		Points: startingPoints,
		Count:  startingCount,
	}
}
