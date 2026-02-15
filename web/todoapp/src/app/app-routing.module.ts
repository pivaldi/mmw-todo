import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { TodoListComponent } from './components/todo-list/todo-list.component';
import { TodoDetailComponent } from './components/todo-detail/todo-detail.component';
import { TodoFormComponent } from './components/todo-form/todo-form.component';

const routes: Routes = [
  { path: '', redirectTo: '/todos', pathMatch: 'full' },
  { path: 'todos', component: TodoListComponent },
  { path: 'todos/new', component: TodoFormComponent },
  { path: 'todos/:id', component: TodoDetailComponent },
  { path: 'todos/:id/edit', component: TodoFormComponent },
  { path: '**', redirectTo: '/todos' }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
