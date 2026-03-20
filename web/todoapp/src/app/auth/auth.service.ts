import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { Router } from '@angular/router';
import { environment } from '../../environments/environment';

const TOKEN_KEY = 'auth_token';
const AUTH_API = `${environment.apiUrl}/auth/auth.v1.AuthService`;
const JSON_HEADERS = new HttpHeaders({ 'Content-Type': 'application/json' });

@Injectable({ providedIn: 'root' })
export class AuthService {
  constructor(private http: HttpClient, private router: Router) {}

  register(login: string, password: string): Observable<{ userId: string }> {
    return this.http.post<{ userId: string }>(
      `${AUTH_API}/Register`,
      { login, password },
      { headers: JSON_HEADERS }
    );
  }

  login(login: string, password: string): Observable<{ token: string; userId: string }> {
    return this.http.post<{ token: string; userId: string }>(
      `${AUTH_API}/Login`,
      { login, password },
      { headers: JSON_HEADERS }
    ).pipe(
      tap(res => localStorage.setItem(TOKEN_KEY, res.token))
    );
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
