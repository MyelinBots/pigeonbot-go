package player

import "fmt"

// Player represents a game player
type Player struct {
	name   string
	points int
	count  int
}

// StartingPoints is the default initial points for players
const StartingPoints = 0

// NewPlayer initializes a new Player instance
func NewPlayer(name string, startingPoints int, startingCount int) *Player {
	return &Player{
		name:   name,
		points: startingPoints,
		count:  startingCount,
	}
}

// Name returns the player's name
func (p *Player) Name() string {
	return p.name
}

// Points returns the player's current points
func (p *Player) Points() int {
	return p.points
}

// ChangePoints modifies the player's points by a specified amount
func (p *Player) ChangePoints(points int) {
	p.points += points
}

// ResetPoints resets the player's points to the starting value
func (p *Player) ResetPoints() {
	p.points = StartingPoints
}

// AddPoints adds a specified amount to the player's points
func (p *Player) AddPoints(points int) {
	p.points += points
}

// RemovePoints removes a specified amount from the player's points
// Ensures the points do not go below zero
func (p *Player) RemovePoints(points int) {
	p.points -= points
	if p.points < 0 {
		p.points = 0
	}
}

// Count returns the player's count value
func (p *Player) Count() int {
	return p.count
}

// AddCount increments the player's count
func (p *Player) AddCount() {
	p.count++
}

// String provides a string representation of the player
func (p *Player) String() string {
	return fmt.Sprintf("%s has %d points.", p.name, p.points)
}
