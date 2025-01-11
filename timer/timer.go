package timer

import (
	"sync"
	"time"
)

// RepeatedTimer provides functionality for running a function at regular intervals
type RepeatedTimer struct {
	interval  time.Duration
	function  func()
	stopChan  chan struct{}
	isRunning bool
	mu        sync.Mutex
}

// NewRepeatedTimer creates and starts a new RepeatedTimer
func NewRepeatedTimer(intervalSeconds int, function func()) *RepeatedTimer {
	rt := &RepeatedTimer{
		interval:  time.Duration(intervalSeconds) * time.Second,
		function:  function,
		stopChan:  make(chan struct{}, 1),
		isRunning: false,
	}

	rt.Start()
	return rt
}

// Start begins the timer
func (rt *RepeatedTimer) Start() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.isRunning {
		rt.isRunning = true
		go rt.run()
	}
}

// run executes the function at the specified interval
func (rt *RepeatedTimer) run() {
	ticker := time.NewTicker(rt.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rt.function()
		case <-rt.stopChan:
			return
		}
	}
}

// Stop halts the timer
func (rt *RepeatedTimer) Stop() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if rt.isRunning {
		rt.isRunning = false
		close(rt.stopChan)
		rt.stopChan = make(chan struct{}, 1) // Create a new stop channel for future restarts
	}
}
