package pigeon

// Pigeon struct represents a pigeon with attributes for type, points, and success rate
type Pigeon struct {
	Type    string
	Points  int
	Success int
}

func NewPigeon(pigeonType string, points, success int) *Pigeon {
	return &Pigeon{
		Type:    pigeonType,
		Points:  points,
		Success: success,
	}
}

func PredefinedPigeons() []*Pigeon {
	return []*Pigeon{
		NewPigeon("cartel member", 10, 85),
		NewPigeon("boss", 100, 25),
		NewPigeon("white", 50, 50),
	}
}
