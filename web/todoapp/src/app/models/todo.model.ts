export enum TaskStatus {
  PENDING = 'pending',
  IN_PROGRESS = 'in_progress',
  COMPLETED = 'completed',
  CANCELLED = 'cancelled'
}

export enum Priority {
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  URGENT = 'urgent'
}

export interface Todo {
  id: string;
  title: string;
  description: string;
  status: TaskStatus;
  priority: Priority;
  dueDate?: Date;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateTodoRequest {
  title: string;
  description: string;
  priority: Priority;
  dueDate?: Date;
}

export interface UpdateTodoRequest {
  title?: string;
  description?: string;
  priority?: Priority;
  dueDate?: Date;
  status?: TaskStatus;
}

export interface ListTodosRequest {
  status?: TaskStatus;
  priority?: Priority;
  limit?: number;
  offset?: number;
}

export interface ListTodosResponse {
  todos: Todo[];
  totalCount: number;
}
