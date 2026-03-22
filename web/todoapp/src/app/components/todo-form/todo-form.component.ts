import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { TodoService } from '../../services/todo.service';
import { create } from '@bufbuild/protobuf';
import { Priority, TaskStatus, CreateTodoRequestSchema, UpdateTodoRequestSchema } from '@contracts/todo/v1/todo_pb';
import { timestampToDate, dateToTimestamp } from '../../pipes/timestamp.pipe';

@Component({
  selector: 'app-todo-form',
  templateUrl: './todo-form.component.html',
  styleUrls: ['./todo-form.component.css']
})
export class TodoFormComponent implements OnInit {
  todoForm: FormGroup;
  isEditMode = false;
  todoId: string | null = null;
  loading = false;
  error: string | null = null;

  // Enums for template
  priorities = [Priority.LOW, Priority.MEDIUM, Priority.HIGH, Priority.URGENT];

  readonly priorityLabels: Record<Priority, string> = {
    [Priority.UNSPECIFIED]: '',
    [Priority.LOW]:    'Low',
    [Priority.MEDIUM]: 'Medium',
    [Priority.HIGH]:   'High',
    [Priority.URGENT]: 'Urgent',
  };

  constructor(
    private fb: FormBuilder,
    private todoService: TodoService,
    private router: Router,
    private route: ActivatedRoute
  ) {
    this.todoForm = this.fb.group({
      title: ['', [Validators.required, Validators.maxLength(200)]],
      description: [''],
      priority: [Priority.MEDIUM, Validators.required],
      dueDate: ['']
    });
  }

  ngOnInit(): void {
    this.todoId = this.route.snapshot.paramMap.get('id');
    this.isEditMode = this.todoId !== null && this.todoId !== 'new';

    if (this.isEditMode && this.todoId) {
      this.loadTodo(this.todoId);
    }
  }

  loadTodo(id: string): void {
    this.loading = true;
    this.error = null;

    this.todoService.getTodo(id).subscribe({
      next: (todo) => {
        this.todoForm.patchValue({
          title: todo.title,
          description: todo.description,
          priority: todo.priority,
          dueDate: todo.dueDate ? this.formatDateForInput(timestampToDate(todo.dueDate)) : ''
        });
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Failed to load todo. Please try again.';
        this.loading = false;
        console.error('Error loading todo:', err);
      }
    });
  }

  onSubmit(): void {
    if (this.todoForm.invalid) {
      this.markFormGroupTouched(this.todoForm);
      return;
    }

    const formValue = this.todoForm.value;
    const fields = {
      title: formValue.title.trim(),
      description: formValue.description?.trim() || '',
      priority: formValue.priority,
      dueDate: formValue.dueDate ? dateToTimestamp(new Date(formValue.dueDate)) : undefined
    };

    this.loading = true;
    this.error = null;

    if (this.isEditMode && this.todoId) {
      const request = create(UpdateTodoRequestSchema, { id: this.todoId, ...fields });
      this.todoService.updateTodo(this.todoId, request).subscribe({
        next: () => {
          this.router.navigate(['/todos', this.todoId]);
        },
        error: (err) => {
          this.error = 'Failed to update todo. Please try again.';
          this.loading = false;
          console.error('Error updating todo:', err);
        }
      });
    } else {
      const request = create(CreateTodoRequestSchema, fields);
      this.todoService.createTodo(request).subscribe({
        next: (todo) => {
          this.router.navigate(['/todos', todo.id]);
        },
        error: (err) => {
          this.error = 'Failed to create todo. Please try again.';
          this.loading = false;
          console.error('Error creating todo:', err);
        }
      });
    }
  }

  cancel(): void {
    if (this.isEditMode && this.todoId) {
      this.router.navigate(['/todos', this.todoId]);
    } else {
      this.router.navigate(['/todos']);
    }
  }

  private formatDateForInput(date: Date | null): string {
    if (!date) return '';
    const d = new Date(date);
    const year  = d.getFullYear();
    const month = String(d.getMonth() + 1).padStart(2, '0');
    const day   = String(d.getDate()).padStart(2, '0');
    const hours = String(d.getHours()).padStart(2, '0');
    const mins  = String(d.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${mins}`;
  }

  private markFormGroupTouched(formGroup: FormGroup): void {
    Object.keys(formGroup.controls).forEach(key => {
      const control = formGroup.get(key);
      control?.markAsTouched();
    });
  }

  // Getter for easy access to form controls in template
  get f() {
    return this.todoForm.controls;
  }
}
