package pigeon_test

import (
	"testing"

	"github.com/MyelinBots/pigeonbot-go/internal/services/pigeon"
	"github.com/stretchr/testify/assert"
)

func TestNewPigeon(t *testing.T) {
	tests := []struct {
		name        string
		pigeonType  string
		points      int
		successRate int
	}{
		{
			name:        "create cartel member pigeon",
			pigeonType:  "cartel member",
			points:      10,
			successRate: 85,
		},
		{
			name:        "create boss pigeon",
			pigeonType:  "boss",
			points:      100,
			successRate: 25,
		},
		{
			name:        "create white pigeon",
			pigeonType:  "white",
			points:      50,
			successRate: 50,
		},
		{
			name:        "create custom pigeon",
			pigeonType:  "custom",
			points:      999,
			successRate: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pigeon.NewPigeon(tt.pigeonType, tt.points, tt.successRate)

			assert.NotNil(t, p)
			assert.Equal(t, tt.pigeonType, p.Type)
			assert.Equal(t, tt.points, p.Points)
			assert.Equal(t, tt.successRate, p.Success)
		})
	}
}

func TestPredefinedPigeons(t *testing.T) {
	pigeons := pigeon.PredefinedPigeons()

	assert.Len(t, pigeons, 3)

	// Test cartel member
	cartelMember := findPigeonByType(pigeons, "cartel member")
	assert.NotNil(t, cartelMember)
	assert.Equal(t, 10, cartelMember.Points)
	assert.Equal(t, 85, cartelMember.Success)

	// Test boss
	boss := findPigeonByType(pigeons, "boss")
	assert.NotNil(t, boss)
	assert.Equal(t, 100, boss.Points)
	assert.Equal(t, 25, boss.Success)

	// Test white
	white := findPigeonByType(pigeons, "white")
	assert.NotNil(t, white)
	assert.Equal(t, 50, white.Points)
	assert.Equal(t, 50, white.Success)
}

func TestPredefinedPigeons_Order(t *testing.T) {
	pigeons := pigeon.PredefinedPigeons()

	// Verify the order is consistent
	assert.Equal(t, "cartel member", pigeons[0].Type)
	assert.Equal(t, "boss", pigeons[1].Type)
	assert.Equal(t, "white", pigeons[2].Type)
}

func TestPredefinedPigeons_SuccessRates(t *testing.T) {
	pigeons := pigeon.PredefinedPigeons()

	// Cartel member should be easiest to hit (85%)
	cartelMember := findPigeonByType(pigeons, "cartel member")
	assert.Equal(t, 85, cartelMember.Success)

	// Boss should be hardest to hit (25%)
	boss := findPigeonByType(pigeons, "boss")
	assert.Equal(t, 25, boss.Success)

	// White should be in the middle (50%)
	white := findPigeonByType(pigeons, "white")
	assert.Equal(t, 50, white.Success)
}

func TestPredefinedPigeons_PointsVsSuccess(t *testing.T) {
	pigeons := pigeon.PredefinedPigeons()

	// Higher points should correlate with lower success rate
	for i := 0; i < len(pigeons); i++ {
		for j := i + 1; j < len(pigeons); j++ {
			if pigeons[i].Points > pigeons[j].Points {
				assert.Less(t, pigeons[i].Success, pigeons[j].Success,
					"Pigeon with more points (%s: %d) should have lower success rate than (%s: %d)",
					pigeons[i].Type, pigeons[i].Points, pigeons[j].Type, pigeons[j].Points)
			}
		}
	}
}

// Helper function to find a pigeon by type
func findPigeonByType(pigeons []*pigeon.Pigeon, pigeonType string) *pigeon.Pigeon {
	for _, p := range pigeons {
		if p.Type == pigeonType {
			return p
		}
	}
	return nil
}
