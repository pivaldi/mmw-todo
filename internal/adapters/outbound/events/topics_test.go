package events

import (
	"testing"

	tododef "github.com/pivaldi/mmw-contracts/go/application/todo"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

func TestDomainTopics_AllEventTypesCovered(t *testing.T) {
	expected := map[string]string{
		domain.EventTypeCreated:   tododef.TopicUserTaskCreated,
		domain.EventTypeUpdated:   tododef.TopicUserTaskUpdated,
		domain.EventTypeCompleted: tododef.TopicUserTaskCompleted,
		domain.EventTypeReopened:  tododef.TopicUserTaskReopened,
		domain.EventTypeDeleted:   tododef.TopicUserTaskDeleted,
	}

	for domainType, wantTopic := range expected {
		t.Run(domainType, func(t *testing.T) {
			got, ok := domainTopics[domainType]
			if !ok {
				t.Fatalf("domainTopics missing entry for %q", domainType)
			}
			if got != wantTopic {
				t.Errorf("domainTopics[%q] = %q, want %q", domainType, got, wantTopic)
			}
		})
	}
}

func TestDomainTopics_TopicIsNotSameAsEventType(t *testing.T) {
	for domainType, topic := range domainTopics {
		if topic == domainType {
			t.Errorf("domainTopics[%q]: topic should be a routing key, not the domain event type", domainType)
		}
	}
}
