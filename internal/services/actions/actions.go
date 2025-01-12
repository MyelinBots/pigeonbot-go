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
