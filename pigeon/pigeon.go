package pigeon

// Pigeon struct represents a pigeon with attributes for type, points, and success rate
type Pigeon struct {
	_type    string
	_points  int
	_success int
}

// NewPigeon initializes a new Pigeon instance
func NewPigeon(pigeonType string, points, success int) *Pigeon {
	return &Pigeon{
		_type:    pigeonType,
		_points:  points,
		_success: success,
	}
}

// Type returns the type of the pigeon
func (p *Pigeon) Type() string {
	return p._type
}

// Points returns the points awarded for this pigeon
func (p *Pigeon) Points() int {
	return p._points
}

// Success returns the success rate of capturing/shooting this pigeon
func (p *Pigeon) Success() int {
	return p._success
}

// PredefinedPigeons returns a list of predefined pigeons
func PredefinedPigeons() []*Pigeon {
	return []*Pigeon{
		NewPigeon("cartel member", 10, 85),
		NewPigeon("boss", 100, 25),
		NewPigeon("white", 50, 50),
	}
}
