package app

import (
	"sync"
	"time"
)

// Item is the CRUD data model.
type Item struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Price   float64 `json:"price"`
	Tags    []string `json:"tags,omitempty"`
	Active  bool    `json:"active"`
}

// DelayedEntry is written by /stream endpoints after a configurable delay.
type DelayedEntry struct {
	ID        int       `json:"id"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

// FailedStatus records every error response served by the app.
type FailedStatus struct {
	Path      string `json:"path"`
	Method    string `json:"method"`
	Status    int    `json:"status"`
	Message   string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	mu   sync.Mutex

	// CRUD store
	items  = make(map[int]Item)
	nextID = 1

	// Delayed output collector
	delayedOutputs []DelayedEntry

	// Rate limiting
	rateTracker = make(map[string]int) // client -> count

	// Conflict tracking (unique names)
	usedNames = make(map[string]bool)

	// Failed status tracker
	failedStatuses []FailedStatus
)

// ResetStore clears all in-memory state. Call between test groups that need
// deterministic state.
func ResetStore() {
	mu.Lock()
	defer mu.Unlock()
	items = make(map[int]Item)
	nextID = 1
	delayedOutputs = nil
	rateTracker = make(map[string]int)
	usedNames = make(map[string]bool)
	failedStatuses = nil
}

func recordFailed(method, path string, status int, msg string) {
	mu.Lock()
	failedStatuses = append(failedStatuses, FailedStatus{
		Path:      path,
		Method:    method,
		Status:    status,
		Message:   msg,
		Timestamp: time.Now(),
	})
	mu.Unlock()
}

func getFailedStatuses() []FailedStatus {
	mu.Lock()
	defer mu.Unlock()
	out := make([]FailedStatus, len(failedStatuses))
	copy(out, failedStatuses)
	return out
}

func checkRateLimit(client string, limit int) bool {
	mu.Lock()
	defer mu.Unlock()
	rateTracker[client]++
	return rateTracker[client] <= limit
}

func isNameUsed(name string) bool {
	mu.Lock()
	defer mu.Unlock()
	if usedNames[name] {
		return true
	}
	usedNames[name] = true
	return false
}

func collectDelayed() []DelayedEntry {
	mu.Lock()
	defer mu.Unlock()
	out := make([]DelayedEntry, len(delayedOutputs))
	copy(out, delayedOutputs)
	return out
}
