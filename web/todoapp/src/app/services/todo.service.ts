import { Injectable } from '@angular/core';
import { Observable, from } from 'rxjs';
import { map } from 'rxjs/operators';
import { createClient, ConnectError, Code } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { TodoService as TodoServiceDef } from '@contracts/todo/v1/todo_pb';
import {
  Todo,
  ListTodosRequest,
  ListTodosResponse,
  CreateTodoRequest,
  UpdateTodoRequest,
} from '@contracts/todo/v1/todo_pb';
import { environment } from '../../environments/environment';

const TOKEN_KEY = 'auth_token';

const transport = createConnectTransport({
  baseUrl: environment.apiUrl,
  interceptors: [
    (next) => async (req) => {
      const token = localStorage.getItem(TOKEN_KEY);
      if (token) req.header.set('Authorization', `Bearer ${token}`);
      try {
        return await next(req);
      } catch (err) {
        if (err instanceof ConnectError && err.code === Code.Unauthenticated) {
          localStorage.removeItem(TOKEN_KEY);
          window.location.href = '/login';
        }
        throw err;
      }
    },
  ],
});

const client = createClient(TodoServiceDef, transport);

@Injectable({ providedIn: 'root' })
export class TodoService {
  listTodos(request?: ListTodosRequest): Observable<ListTodosResponse> {
    return from(client.listTodos(request ?? {}));
  }

  getTodo(id: string): Observable<Todo> {
    return from(client.getTodo({ id })).pipe(map((res) => res.todo!));
  }

  createTodo(request: CreateTodoRequest): Observable<Todo> {
    return from(client.createTodo(request)).pipe(map((res) => res.todo!));
  }

  updateTodo(id: string, request: Omit<UpdateTodoRequest, 'id'>): Observable<Todo> {
    return from(client.updateTodo({ id, ...request })).pipe(map((res) => res.todo!));
  }

  completeTodo(id: string): Observable<Todo> {
    return from(client.completeTodo({ id })).pipe(map((res) => res.todo!));
  }

  reopenTodo(id: string): Observable<Todo> {
    return from(client.reopenTodo({ id })).pipe(map((res) => res.todo!));
  }

  deleteTodo(id: string): Observable<void> {
    return from(client.deleteTodo({ id })).pipe(map(() => void 0));
  }
}
