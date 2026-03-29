import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ConnectError } from '@connectrpc/connect';
import { AuthErrorCode } from '@contracts/auth/v1/auth_pb';
import { DomainErrorSchema } from '@contracts/common/v1/errors_pb';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html'
})
export class RegisterComponent {
  login = '';
  password = '';
  errorMessage = '';

  constructor(private auth: AuthService, private router: Router) {}

  onSubmit(): void {
    this.errorMessage = '';
    this.auth.register(this.login, this.password).subscribe({
      next: () => this.router.navigate(['/login']),
      error: (err: unknown) => {
        this.errorMessage = this.getRegisterError(err);
      }
    });
  }

  private getRegisterError(err: unknown): string {
    if (!(err instanceof ConnectError)) {
      return 'An unexpected error occurred. Please try again.';
    }

    const detail = err.findDetails(DomainErrorSchema)[0];

    switch (detail?.code) {
      case AuthErrorCode.INVALID_LOGIN:
        return 'Login must not be empty.';
      case AuthErrorCode.INVALID_PASSWORD:
        return 'Password must not be empty.';
      case AuthErrorCode.USER_ALREADY_EXISTS:
        return 'This login is already taken.';
      default:
        return 'Registration failed. Please try again.';
    }
  }
}
