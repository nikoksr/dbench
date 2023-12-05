package events

import "sync"

// EventType represents the type of event.
type EventType string

// Event represents an event with a type and a message.
type Event struct {
	Type    EventType // The type of the event.
	Message string    // The message of the event.
}

// EventHandler represents a function that handles an event.
type EventHandler func(Event)

var (
	mu       sync.Mutex     // Mutex for thread-safe operations.
	handlers []EventHandler // Slice of event handlers.
)

// Subscribe adds a new event handler to the handlers slice.
// It locks the mutex before appending to the slice and unlocks it after.
func Subscribe(handler EventHandler) {
	mu.Lock()
	defer mu.Unlock()
	handlers = append(handlers, handler)
}

// PublishEvent calls each handler with the provided event.
// It locks the mutex before iterating over the handlers and unlocks it after.
// Each handler is called in a new goroutine for non-blocking event notifications.
func PublishEvent(event Event) {
	mu.Lock()
	defer mu.Unlock()
	for _, handler := range handlers {
		go handler(event)
	}
}
