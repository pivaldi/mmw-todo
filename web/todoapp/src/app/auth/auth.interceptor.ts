import { Injectable } from '@angular/core';
import {
  HttpInterceptor, HttpRequest, HttpHandler,
  HttpEvent, HttpErrorResponse
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { AuthService } from './auth.service';

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private auth: AuthService) {}

  intercept(req: HttpRequest<unknown>, next: HttpHandler): Observable<HttpEvent<unknown>> {
    // Only attach token for /api/ requests — leave third-party and asset requests alone
    if (!req.url.startsWith('/api/')) {
      return next.handle(req);
    }

    const token = this.auth.getToken();
    const authedReq = token
      ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
      : req;

    return next.handle(authedReq).pipe(
      catchError((err: HttpErrorResponse) => {
        if (err.status === 401) {
          // Expired or invalid token — clear localStorage and redirect to login
          this.auth.logout();
        }

        return throwError(() => err);
      })
    );
  }
}
