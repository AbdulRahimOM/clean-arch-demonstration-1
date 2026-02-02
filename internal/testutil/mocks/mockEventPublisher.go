package mocks

import (
	"context"
)

// MockEventPublisher implements interfaces.EventPublisher for tests.
// Published records all Publish calls for assertions.
type MockEventPublisher struct {
	Published  []interface{}
	PublishErr error
}

func (m *MockEventPublisher) Publish(ctx context.Context, event interface{}) error {
	m.Published = append(m.Published, event)
	return m.PublishErr
}
