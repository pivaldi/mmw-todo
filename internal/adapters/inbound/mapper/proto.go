// modules/todo/internal/adapters/inbound/mapper/proto.go
package mapper

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	todov1 "github.com/pivaldi/mmw-contracts/go/network/todo/v1"
	"github.com/pivaldi/mmw-todo/internal/application/dto"
	"github.com/pivaldi/mmw-todo/internal/domain"
)

func TodoToProto(todo *dto.TodoResponse) *todov1.Todo {
	protoTodo := &todov1.Todo{
		Id:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Status:      StatusToProto(todo.Status),
		Priority:    PriorityToProto(todo.Priority),
		CreatedAt:   timestamppb.New(todo.CreatedAt),
		UpdatedAt:   timestamppb.New(todo.UpdatedAt),
	}

	if todo.DueDate != nil {
		protoTodo.DueDate = timestamppb.New(*todo.DueDate)
	}

	return protoTodo
}

func StatusToProto(status domain.TaskStatus) todov1.TaskStatus {
	switch status {
	case domain.TaskStatusPending:
		return todov1.TaskStatus_TASK_STATUS_PENDING
	case domain.TaskStatusInProgress:
		return todov1.TaskStatus_TASK_STATUS_IN_PROGRESS
	case domain.TaskStatusCompleted:
		return todov1.TaskStatus_TASK_STATUS_COMPLETED
	case domain.TaskStatusCancelled:
		return todov1.TaskStatus_TASK_STATUS_CANCELLED
	default:
		return todov1.TaskStatus_TASK_STATUS_UNSPECIFIED
	}
}

func PriorityToProto(priority domain.Priority) todov1.Priority {
	switch priority {
	case domain.PriorityLow:
		return todov1.Priority_PRIORITY_LOW
	case domain.PriorityMedium:
		return todov1.Priority_PRIORITY_MEDIUM
	case domain.PriorityHigh:
		return todov1.Priority_PRIORITY_HIGH
	case domain.PriorityUrgent:
		return todov1.Priority_PRIORITY_URGENT
	default:
		return todov1.Priority_PRIORITY_UNSPECIFIED
	}
}

func StatusFromProto(status todov1.TaskStatus) domain.TaskStatus {
	switch status {
	case todov1.TaskStatus_TASK_STATUS_IN_PROGRESS:
		return domain.TaskStatusInProgress
	case todov1.TaskStatus_TASK_STATUS_COMPLETED:
		return domain.TaskStatusCompleted
	case todov1.TaskStatus_TASK_STATUS_CANCELLED:
		return domain.TaskStatusCancelled
	default:
		return domain.TaskStatusPending
	}
}

func PriorityFromProto(priority todov1.Priority) domain.Priority {
	switch priority {
	case todov1.Priority_PRIORITY_LOW:
		return domain.PriorityLow
	case todov1.Priority_PRIORITY_HIGH:
		return domain.PriorityHigh
	case todov1.Priority_PRIORITY_URGENT:
		return domain.PriorityUrgent
	default:
		return domain.PriorityMedium
	}
}
