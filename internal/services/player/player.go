package player

import "fmt"

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

// String provides a string representation of the player
func (p *Player) String() string {
	return fmt.Sprintf("%s has %d points.", p.Name, p.Points)
}

// Helper function to determine the player's level
func (p *Player) GetPlayerLevel() string {
	switch {
	case p.Count >= 10 && p.Count <= 100:
		return "Initiate 🐦"
	case p.Count >= 101 && p.Count <= 199:
		return "Adept 🦅"
	case p.Count >= 200 && p.Count <= 499:
		return "Expert 🕊️"
	case p.Count >= 500 && p.Count <= 799:
		return "Master 🦜"
	case p.Count >= 800 && p.Count <= 999:
		return "Grandmaster 🐔"
	case p.Count >= 1000 && p.Count <= 2999:
		return "Legendary Phoenix 🐉🔥"
	case p.Count >= 3000 && p.Count <= 4999:
		return "Mythic Dragon 🐲✨"
	case p.Count >= 5000 && p.Count <= 9999:
		return "Cosmic Falcon 🌌🦅"
	case p.Count >= 10000:
		return "Lord of Pigeons 👑🐦"
	default:
		return "Beginner 🐣"
	}
}
