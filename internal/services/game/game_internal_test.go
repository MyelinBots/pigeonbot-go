package game

import (
	"testing"

	"github.com/MyelinBots/pigeonbot-go/internal/services/actions"
	"github.com/stretchr/testify/assert"
)

func TestMedal(t *testing.T) {
	tests := []struct {
		rank     int
		expected string
	}{
		{0, "🥇"},
		{1, "🥈"},
		{2, "🥉"},
		{3, "•"},
		{4, "•"},
		{10, "•"},
		{100, "•"},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.rank)), func(t *testing.T) {
			result := medal(tt.rank)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorFunction(t *testing.T) {
	// Test the color formatting function
	result := c("test", 8)
	// Should contain IRC color codes
	assert.Contains(t, result, ircColor)
	assert.Contains(t, result, "test")
	assert.Contains(t, result, ircReset)
}

func TestIRCConstants(t *testing.T) {
	assert.Equal(t, "\x02", ircBold)
	assert.Equal(t, "\x03", ircColor)
	assert.Equal(t, "\x0F", ircReset)
}

func TestPredefinedActions(t *testing.T) {
	actions := predefinedActions()

	assert.Len(t, actions, 4)

	// Check stole action
	stoleAction := findActionByName(actions, "stole")
	assert.NotNil(t, stoleAction)
	assert.Equal(t, 10, stoleAction.ActionPoint)
	assert.Len(t, stoleAction.Items, 3)
	assert.Contains(t, stoleAction.Format, "%s")

	// Check pooped action
	poopedAction := findActionByName(actions, "pooped")
	assert.NotNil(t, poopedAction)
	assert.Equal(t, 10, poopedAction.ActionPoint)
	assert.Len(t, poopedAction.Items, 3)

	// Check landed action
	landedAction := findActionByName(actions, "landed")
	assert.NotNil(t, landedAction)
	assert.Equal(t, 10, landedAction.ActionPoint)
	assert.Len(t, landedAction.Items, 8)

	// Check mating action
	matingAction := findActionByName(actions, "mating")
	assert.NotNil(t, matingAction)
	assert.Equal(t, 10, matingAction.ActionPoint)
	assert.Len(t, matingAction.Items, 6)
}

func findActionByName(actionsList []actions.Action, name string) *actions.Action {
	for i := range actionsList {
		if actionsList[i].Action == name {
			return &actionsList[i]
		}
	}
	return nil
}
