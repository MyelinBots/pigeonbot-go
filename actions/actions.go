package actions

import (
	"fmt"
	"math/rand"
	"time"
)

// Action represents an action with associated data
type Action struct {
	Action      string
	Items       []string
	Format      string
	ActionPoint int
}

// Act performs the action and returns the formatted result
func (a *Action) Act(name string) string {
	rand.Seed(time.Now().UnixNano())
	item := a.Items[rand.Intn(len(a.Items))]
	return fmt.Sprintf(a.Format, name, a.Action, item)
}

// Doactions initializes and performs some example actions
func Doactions() {
	action := Action{
		Action:      "play",
		Items:       []string{"ball", "frisbee", "rope"},
		Format:      "%s wants to %s with a %s.",
		ActionPoint: 10,
	}

	result := action.Act("Alice")
	fmt.Println(result)
}
