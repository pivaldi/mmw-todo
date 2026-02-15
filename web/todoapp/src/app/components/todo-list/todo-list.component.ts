import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { TodoService } from '../../services/todo.service';
import { Todo, TaskStatus, Priority } from '../../models/todo.model';

@Component({
  selector: 'app-todo-list',
  templateUrl: './todo-list.component.html',
  styleUrls: ['./todo-list.component.css']
})
export class TodoListComponent implements OnInit {
  todos: Todo[] = [];
  filteredTodos: Todo[] = [];
  loading = false;
  error: string | null = null;

  // Filter options
  statusFilter: TaskStatus | 'all' = 'all';
  priorityFilter: Priority | 'all' = 'all';

  // Enums for template
  TaskStatus = TaskStatus;
  Priority = Priority;

  constructor(
    private todoService: TodoService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadTodos();
  }

  loadTodos(): void {
    this.loading = true;
    this.error = null;

    this.todoService.listTodos().subscribe({
      next: (response) => {
        this.todos = response.todos;
        this.applyFilters();
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load todos. Please try again.';
        this.loading = false;
        console.error('Error loading todos:', err);
      }
    });
  }

  applyFilters(): void {
    this.filteredTodos = this.todos.filter(todo => {
      const statusMatch = this.statusFilter === 'all' || todo.status === this.statusFilter;
      const priorityMatch = this.priorityFilter === 'all' || todo.priority === this.priorityFilter;
      return statusMatch && priorityMatch;
    });
  }

  onFilterChange(): void {
    this.applyFilters();
  }

  viewTodo(id: string): void {
    this.router.navigate(['/todos', id]);
  }

  createTodo(): void {
    this.router.navigate(['/todos/new']);
  }

  completeTodo(todo: Todo, event: Event): void {
    event.stopPropagation();

    if (todo.status === TaskStatus.COMPLETED) {
      return;
    }

    this.todoService.completeTodo(todo.id).subscribe({
      next: (updatedTodo) => {
        const index = this.todos.findIndex(t => t.id === todo.id);
        if (index !== -1) {
          this.todos[index] = updatedTodo;
          this.applyFilters();
        }
      },
      error: (err) => {
        console.error('Error completing todo:', err);
        alert('Failed to complete todo');
      }
    });
  }

  deleteTodo(todo: Todo, event: Event): void {
    event.stopPropagation();

    if (!confirm(`Are you sure you want to delete "${todo.title}"?`)) {
      return;
    }

    this.todoService.deleteTodo(todo.id).subscribe({
      next: () => {
        this.todos = this.todos.filter(t => t.id !== todo.id);
        this.applyFilters();
      },
      error: (err) => {
        console.error('Error deleting todo:', err);
        alert('Failed to delete todo');
      }
    });
  }

  getStatusClass(status: TaskStatus): string {
    const classes: Record<TaskStatus, string> = {
      [TaskStatus.PENDING]: 'status-pending',
      [TaskStatus.IN_PROGRESS]: 'status-in-progress',
      [TaskStatus.COMPLETED]: 'status-completed',
      [TaskStatus.CANCELLED]: 'status-cancelled'
    };
    return classes[status] || '';
  }

  getPriorityClass(priority: Priority): string {
    const classes: Record<Priority, string> = {
      [Priority.LOW]: 'priority-low',
      [Priority.MEDIUM]: 'priority-medium',
      [Priority.HIGH]: 'priority-high',
      [Priority.URGENT]: 'priority-urgent'
    };
    return classes[priority] || '';
  }

  isDue(todo: Todo): boolean {
    if (!todo.dueDate) {
      return false;
    }
    return new Date(todo.dueDate) < new Date();
  }

  isDueSoon(todo: Todo): boolean {
    if (!todo.dueDate) {
      return false;
    }
    const now = new Date();
    const dueDate = new Date(todo.dueDate);
    const hoursDiff = (dueDate.getTime() - now.getTime()) / (1000 * 60 * 60);
    return hoursDiff > 0 && hoursDiff <= 24;
  }
}
