package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRareEggConstants(t *testing.T) {
	// Verify constants are set correctly
	assert.Equal(t, 10, rareEggAppearPercent, "Rare egg appear chance should be 10%")
	assert.Equal(t, 50, rareEggSuccessPercent, "Rare egg success chance should be 50%")
	assert.Equal(t, 10000, rareEggPointBoost, "Rare egg point boost should be 10000")
}
