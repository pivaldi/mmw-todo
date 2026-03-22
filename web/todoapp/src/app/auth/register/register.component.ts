import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { ConnectError, Code } from '@connectrpc/connect';
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
        this.errorMessage = err instanceof ConnectError && err.code === Code.AlreadyExists
          ? 'This login is already taken'
          : 'Registration failed. Please try again.';
      }
    });
  }
}
