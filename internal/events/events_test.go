package events

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testEvents = []struct {
	typ     EventType
	message string
}{
	{typ: "Test", message: "This is Test Event"},
	{typ: "", message: ""},
}

func TestSubscribe(t *testing.T) {
	Subscribe(func(Event) {})

	mu.Lock()
	handlerCount := len(handlers)
	mu.Unlock()

	assert.Equal(t, 1, handlerCount, "Expected one handler in the handlers slice")
}

func TestPublishEvent(t *testing.T) {
	eventCh := make(chan Event, 1)

	Subscribe(func(event Event) {
		eventCh <- event
	})

	for _, tc := range testEvents {
		event := Event{Type: tc.typ, Message: tc.message}

		go PublishEvent(event)

		select {
		case receivedEvent := <-eventCh:
			assert.Equal(t, event.Type, receivedEvent.Type, "Event types do not match")
			assert.Equal(t, event.Message, receivedEvent.Message, "Event messages do not match")
		case <-time.After(3 * time.Second):
			t.Fatal("Test timed out waiting for event")
		}
	}
}

func TestPublishEventWithNoSubscribers(t *testing.T) {
	mu.Lock()
	handlers = nil
	mu.Unlock()

	assert.NotPanics(t, func() { PublishEvent(Event{}) }, "PublishEvent should not panic with no subscribers")
}
