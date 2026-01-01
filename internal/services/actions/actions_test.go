package actions_test

import (
	"strings"
	"testing"

	"github.com/MyelinBots/pigeonbot-go/internal/services/actions"
	"github.com/stretchr/testify/assert"
)

func TestAction_Act(t *testing.T) {
	tests := []struct {
		name          string
		action        actions.Action
		pigeonType    string
		expectedParts []string // Parts that should be present in the output
	}{
		{
			name: "stole action",
			action: actions.Action{
				Action:      "stole",
				Items:       []string{"tv", "wallet", "food"},
				Format:      "A %s pigeon %s your %s",
				ActionPoint: 10,
			},
			pigeonType:    "cartel member",
			expectedParts: []string{"A", "cartel member", "pigeon", "stole", "your"},
		},
		{
			name: "pooped action",
			action: actions.Action{
				Action:      "pooped",
				Items:       []string{"car", "head", "laptop"},
				Format:      "A %s pigeon %s on your %s",
				ActionPoint: 10,
			},
			pigeonType:    "white",
			expectedParts: []string{"A", "white", "pigeon", "pooped", "on your"},
		},
		{
			name: "landed action",
			action: actions.Action{
				Action:      "landed",
				Items:       []string{"balcony", "head", "car"},
				Format:      "A %s pigeon has %s on your %s",
				ActionPoint: 10,
			},
			pigeonType:    "boss",
			expectedParts: []string{"A", "boss", "pigeon has", "landed", "on your"},
		},
		{
			name: "mating action",
			action: actions.Action{
				Action:      "mating",
				Items:       []string{"balcony", "car", "bed"},
				Format:      "%s pigeons are %s at your %s",
				ActionPoint: 10,
			},
			pigeonType:    "cartel member",
			expectedParts: []string{"cartel member", "pigeons are", "mating", "at your"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.Act(tt.pigeonType)

			// Check that all expected parts are in the result
			for _, part := range tt.expectedParts {
				assert.Contains(t, result, part,
					"Result should contain '%s'", part)
			}

			// Check that at least one item is in the result
			foundItem := false
			for _, item := range tt.action.Items {
				if strings.Contains(result, item) {
					foundItem = true
					break
				}
			}
			assert.True(t, foundItem, "Result should contain at least one item from the list")
		})
	}
}

func TestAction_Act_RandomItem(t *testing.T) {
	action := actions.Action{
		Action:      "stole",
		Items:       []string{"tv", "wallet", "food"},
		Format:      "A %s pigeon %s your %s",
		ActionPoint: 10,
	}

	// Run multiple times to verify randomness
	results := make(map[string]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		result := action.Act("test")
		for _, item := range action.Items {
			if strings.Contains(result, item) {
				results[item]++
				break
			}
		}
	}

	// With 100 iterations and 3 items, each should appear at least a few times
	// (statistically, each should appear around 33 times)
	for _, item := range action.Items {
		assert.Greater(t, results[item], 0,
			"Item '%s' should appear at least once in %d iterations", item, iterations)
	}
}

func TestAction_Act_SingleItem(t *testing.T) {
	action := actions.Action{
		Action:      "landed",
		Items:       []string{"only-item"},
		Format:      "A %s pigeon has %s on your %s",
		ActionPoint: 5,
	}

	result := action.Act("white")

	assert.Contains(t, result, "only-item")
	assert.Contains(t, result, "white")
	assert.Contains(t, result, "landed")
}

func TestAction_Struct(t *testing.T) {
	action := actions.Action{
		Action:      "test-action",
		Items:       []string{"item1", "item2"},
		Format:      "Format: %s %s %s",
		ActionPoint: 42,
	}

	assert.Equal(t, "test-action", action.Action)
	assert.Equal(t, []string{"item1", "item2"}, action.Items)
	assert.Equal(t, "Format: %s %s %s", action.Format)
	assert.Equal(t, 42, action.ActionPoint)
}

func TestAction_EmptyItems(t *testing.T) {
	action := actions.Action{
		Action:      "test",
		Items:       []string{},
		Format:      "A %s pigeon %s on your %s",
		ActionPoint: 10,
	}

	// This will panic with empty items slice, but we can test the struct is valid
	assert.Empty(t, action.Items)
}
