package domain_test

import (
	"testing"

	"github.com/pivaldi/mmw-todo/internal/domain"
)

func TestEventTypes_ReturnOwnConstants(t *testing.T) {
	cases := []struct {
		event domain.DomainEvent
		want  string
	}{
		{&domain.TodoCreated{}, domain.EventTypeCreated},
		{&domain.TodoUpdated{}, domain.EventTypeUpdated},
		{&domain.TodoCompleted{}, domain.EventTypeCompleted},
		{&domain.TodoReopened{}, domain.EventTypeReopened},
		{&domain.TodoDeleted{}, domain.EventTypeDeleted},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.event.EventType(); got != tc.want {
				t.Errorf("%T.EventType() = %q, want %q", tc.event, got, tc.want)
			}
		})
	}
}
