import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable, from } from 'rxjs';
import { tap } from 'rxjs/operators';
import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { AuthPublicService } from '@contracts/auth/v1/auth_pb';
import type { LoginResponse, RegisterResponse } from '@contracts/auth/v1/auth_pb';
import { environment } from '../../environments/environment';

const TOKEN_KEY = 'auth_token';

const transport = createConnectTransport({ baseUrl: environment.apiUrl });
const client = createClient(AuthPublicService, transport);

@Injectable({ providedIn: 'root' })
export class AuthService {
  constructor(private router: Router) {}

  register(login: string, password: string): Observable<RegisterResponse> {
    return from(client.register({ login, password }));
  }

  login(login: string, password: string): Observable<LoginResponse> {
    return from(client.login({ login, password })).pipe(tap((res) => localStorage.setItem(TOKEN_KEY, res.token)));
  }

  logout(): void {
    localStorage.removeItem(TOKEN_KEY);
    this.router.navigate(['/login']);
  }

  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  }

  isLoggedIn(): boolean {
    return this.getToken() !== null;
  }
}
