import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { TodoService } from '../../services/todo.service';
import { Todo, TaskStatus, Priority } from '@contracts/todo/v1/todo_pb';
import { timestampToDate } from '../../pipes/timestamp.pipe';

@Component({
  selector: 'app-todo-detail',
  templateUrl: './todo-detail.component.html',
  styleUrls: ['./todo-detail.component.css']
})
export class TodoDetailComponent implements OnInit {
  todo: Todo | null = null;
  loading = false;
  error: string | null = null;

  // Enums for template
  TaskStatus = TaskStatus;
  Priority = Priority;

  readonly statusLabels: Record<TaskStatus, string> = {
    [TaskStatus.UNSPECIFIED]: '',
    [TaskStatus.PENDING]:     'Pending',
    [TaskStatus.IN_PROGRESS]: 'In Progress',
    [TaskStatus.COMPLETED]:   'Completed',
    [TaskStatus.CANCELLED]:   'Cancelled',
  };

  readonly priorityLabels: Record<Priority, string> = {
    [Priority.UNSPECIFIED]: '',
    [Priority.LOW]:    'Low',
    [Priority.MEDIUM]: 'Medium',
    [Priority.HIGH]:   'High',
    [Priority.URGENT]: 'Urgent',
  };

  constructor(
    private todoService: TodoService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.loadTodo(id);
    }
  }

  loadTodo(id: string): void {
    this.loading = true;
    this.error = null;

    this.todoService.getTodo(id).subscribe({
      next: (todo) => {
        this.todo = todo;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load todo. Please try again.';
        this.loading = false;
        console.error('Error loading todo:', err);
      }
    });
  }

  editTodo(): void {
    if (this.todo) {
      this.router.navigate(['/todos', this.todo.id, 'edit']);
    }
  }

  completeTodo(): void {
    if (!this.todo || this.todo.status === TaskStatus.COMPLETED) {
      return;
    }

    this.todoService.completeTodo(this.todo.id).subscribe({
      next: (updatedTodo) => {
        this.todo = updatedTodo;
      },
      error: (err) => {
        console.error('Error completing todo:', err);
        alert('Failed to complete todo');
      }
    });
  }

  reopenTodo(): void {
    if (!this.todo) {
      return;
    }

    this.todoService.reopenTodo(this.todo.id).subscribe({
      next: (updatedTodo) => {
        this.todo = updatedTodo;
      },
      error: (err) => {
        console.error('Error reopening todo:', err);
        alert('Failed to reopen todo');
      }
    });
  }

  deleteTodo(): void {
    if (!this.todo) {
      return;
    }

    if (!confirm(`Are you sure you want to delete "${this.todo.title}"?`)) {
      return;
    }

    this.todoService.deleteTodo(this.todo.id).subscribe({
      next: () => {
        this.router.navigate(['/todos']);
      },
      error: (err) => {
        console.error('Error deleting todo:', err);
        alert('Failed to delete todo');
      }
    });
  }

  goBack(): void {
    this.router.navigate(['/todos']);
  }

  getStatusClass(status: TaskStatus): string {
    const classes: Record<TaskStatus, string> = {
      [TaskStatus.UNSPECIFIED]: '',
      [TaskStatus.PENDING]:     'status-pending',
      [TaskStatus.IN_PROGRESS]: 'status-in-progress',
      [TaskStatus.COMPLETED]:   'status-completed',
      [TaskStatus.CANCELLED]:   'status-cancelled',
    };
    return classes[status] ?? '';
  }

  getPriorityClass(priority: Priority): string {
    const classes: Record<Priority, string> = {
      [Priority.UNSPECIFIED]: '',
      [Priority.LOW]:    'priority-low',
      [Priority.MEDIUM]: 'priority-medium',
      [Priority.HIGH]:   'priority-high',
      [Priority.URGENT]: 'priority-urgent',
    };
    return classes[priority] ?? '';
  }

  isDue(): boolean {
    const d = timestampToDate(this.todo?.dueDate);
    return d !== null && d < new Date();
  }

  isDueSoon(): boolean {
    const d = timestampToDate(this.todo?.dueDate);
    if (!d) return false;
    const hoursDiff = (d.getTime() - Date.now()) / (1000 * 60 * 60);
    return hoursDiff > 0 && hoursDiff <= 24;
  }
}
