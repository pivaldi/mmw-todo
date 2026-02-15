import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../environments/environment';
import {
  Todo,
  CreateTodoRequest,
  UpdateTodoRequest,
  ListTodosRequest,
  ListTodosResponse
} from '../models/todo.model';

@Injectable({
  providedIn: 'root'
})
export class TodoService {
  private apiUrl = `${environment.apiUrl}/todo.v1.TodoService`;
  private headers = new HttpHeaders({
    'Content-Type': 'application/json'
  });

  constructor(private http: HttpClient) {}

  /**
   * Get all todos with optional filters
   */
  listTodos(request?: ListTodosRequest): Observable<ListTodosResponse> {
    // Connect protocol uses POST for all operations
    const body = request || {};

    return this.http.post<ListTodosResponse>(`${this.apiUrl}/ListTodos`, body, { headers: this.headers }).pipe(
      map(response => ({
        todos: response.todos.map(todo => this.convertDates(todo)),
        totalCount: response.totalCount
      }))
    );
  }

  /**
   * Get a single todo by ID
   */
  getTodo(id: string): Observable<Todo> {
    return this.http.post<{ todo: Todo }>(`${this.apiUrl}/GetTodo`, { id }, { headers: this.headers }).pipe(
      map(response => this.convertDates(response.todo))
    );
  }

  /**
   * Create a new todo
   */
  createTodo(request: CreateTodoRequest): Observable<Todo> {
    return this.http.post<{ todo: Todo }>(`${this.apiUrl}/CreateTodo`, request, { headers: this.headers }).pipe(
      map(response => this.convertDates(response.todo))
    );
  }

  /**
   * Update an existing todo
   */
  updateTodo(id: string, request: UpdateTodoRequest): Observable<Todo> {
    return this.http.post<{ todo: Todo }>(`${this.apiUrl}/UpdateTodo`, {
      id,
      ...request
    }, { headers: this.headers }).pipe(
      map(response => this.convertDates(response.todo))
    );
  }

  /**
   * Mark a todo as completed
   */
  completeTodo(id: string): Observable<Todo> {
    return this.http.post<{ todo: Todo }>(`${this.apiUrl}/CompleteTodo`, { id }, { headers: this.headers }).pipe(
      map(response => this.convertDates(response.todo))
    );
  }

  /**
   * Reopen a completed/cancelled todo
   */
  reopenTodo(id: string): Observable<Todo> {
    return this.http.post<{ todo: Todo }>(`${this.apiUrl}/ReopenTodo`, { id }, { headers: this.headers }).pipe(
      map(response => this.convertDates(response.todo))
    );
  }

  /**
   * Delete a todo
   */
  deleteTodo(id: string): Observable<void> {
    return this.http.post<void>(`${this.apiUrl}/DeleteTodo`, { id }, { headers: this.headers });
  }

  /**
   * Convert string dates to Date objects
   */
  private convertDates(todo: any): Todo {
    return {
      ...todo,
      dueDate: todo.dueDate ? new Date(todo.dueDate) : undefined,
      createdAt: new Date(todo.createdAt),
      updatedAt: new Date(todo.updatedAt)
    };
  }
}
