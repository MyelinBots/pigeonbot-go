package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseEggsByType(t *testing.T) {
	tests := []struct {
		name       string
		pigeonType string
		expected   int
	}{
		{
			name:       "cartel member",
			pigeonType: "cartel member",
			expected:   1,
		},
		{
			name:       "cartel member uppercase",
			pigeonType: "CARTEL MEMBER",
			expected:   1,
		},
		{
			name:       "cartel member mixed case",
			pigeonType: "Cartel Member",
			expected:   1,
		},
		{
			name:       "white",
			pigeonType: "white",
			expected:   2,
		},
		{
			name:       "white uppercase",
			pigeonType: "WHITE",
			expected:   2,
		},
		{
			name:       "boss",
			pigeonType: "boss",
			expected:   5,
		},
		{
			name:       "boss uppercase",
			pigeonType: "BOSS",
			expected:   5,
		},
		{
			name:       "unknown type",
			pigeonType: "unknown",
			expected:   0,
		},
		{
			name:       "empty type",
			pigeonType: "",
			expected:   0,
		},
		{
			name:       "type with leading space",
			pigeonType: " white",
			expected:   2,
		},
		{
			name:       "type with trailing space",
			pigeonType: "white ",
			expected:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := baseEggsByType(tt.pigeonType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEggsAfterCrack_CartelMember(t *testing.T) {
	// Cartel member always cracks all eggs (1 -> 0)
	for i := 0; i < 10; i++ {
		final, cracked := eggsAfterCrack("cartel member")
		assert.Equal(t, 0, final, "Cartel member should always have 0 eggs")
		assert.Equal(t, 1, cracked, "Cartel member should always crack 1 egg")
	}
}

func TestEggsAfterCrack_White(t *testing.T) {
	// White can have 0 or 1 final eggs (from base of 2)
	results := make(map[int]int)

	for i := 0; i < 100; i++ {
		final, cracked := eggsAfterCrack("white")
		assert.True(t, final >= 0 && final <= 1,
			"White pigeon final eggs should be 0 or 1, got %d", final)
		assert.Equal(t, 2-final, cracked,
			"Cracked should equal base - final")
		results[final]++
	}

	// With 100 iterations, we should see both 0 and 1
	assert.Greater(t, results[0], 0, "Should see some 0 results")
	assert.Greater(t, results[1], 0, "Should see some 1 results")
}

func TestEggsAfterCrack_Boss(t *testing.T) {
	// Boss can have 0-4 final eggs (from base of 5)
	results := make(map[int]int)

	for i := 0; i < 200; i++ {
		final, cracked := eggsAfterCrack("boss")
		assert.True(t, final >= 0 && final <= 4,
			"Boss pigeon final eggs should be 0-4, got %d", final)
		assert.Equal(t, 5-final, cracked,
			"Cracked should equal base - final")
		results[final]++
	}

	// With 200 iterations, we should see most values
	// (statistically, each should appear around 40 times)
	for i := 0; i <= 4; i++ {
		assert.Greater(t, results[i], 0,
			"Should see some results with %d eggs", i)
	}
}

func TestEggsAfterCrack_Unknown(t *testing.T) {
	final, cracked := eggsAfterCrack("unknown")
	assert.Equal(t, 0, final)
	assert.Equal(t, 0, cracked)
}

func TestEggsAfterCrack_Empty(t *testing.T) {
	final, cracked := eggsAfterCrack("")
	assert.Equal(t, 0, final)
	assert.Equal(t, 0, cracked)
}

func TestEggsAfterCrack_CaseInsensitive(t *testing.T) {
	// Test case insensitivity
	types := []string{"CARTEL MEMBER", "Cartel Member", "cartel member"}
	for _, pigeonType := range types {
		final, cracked := eggsAfterCrack(pigeonType)
		assert.Equal(t, 0, final)
		assert.Equal(t, 1, cracked)
	}

	// Boss case insensitivity
	for i := 0; i < 20; i++ {
		final1, _ := eggsAfterCrack("BOSS")
		assert.True(t, final1 >= 0 && final1 <= 4)

		final2, _ := eggsAfterCrack("Boss")
		assert.True(t, final2 >= 0 && final2 <= 4)
	}
}
