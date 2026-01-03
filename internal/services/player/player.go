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
	case p.Count < 10:
		return "Beginner 🐣"
	case p.Count <= 100:
		return "Initiate 🐦"
	case p.Count <= 199:
		return "Adept 🦅"
	case p.Count <= 499:
		return "Expert 🕊️"
	case p.Count <= 799:
		return "Master 🦜"
	case p.Count <= 999:
		return "Grandmaster 🐔"
	case p.Count <= 2999:
		return "Legendary Phoenix 🐉🔥"
	case p.Count <= 4999:
		return "Mythic Dragon 🐲✨"
	case p.Count <= 9999:
		return "Cosmic Falcon 🌌🦅"
	case p.Count <= 14999:
		return "Lord of Pigeons 👑🐦"
	case p.Count <= 24999:
		return "Pigeon Emperor 🏯🐦"
	case p.Count <= 39999:
		return "Sky Tyrant ☁️🐲"
	case p.Count <= 59999:
		return "Celestial Hunter 🌠🦅"
	case p.Count <= 99999:
		return "Eternal Wing 🕊️♾️"
	default:
		return "Pigeon God ☄️👁️"
	}
}
