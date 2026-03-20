import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html'
})
export class LoginComponent {
  login = '';
  password = '';
  errorMessage = '';

  constructor(private auth: AuthService, private router: Router) {}

  onSubmit(): void {
    this.errorMessage = '';
    this.auth.login(this.login, this.password).subscribe({
      next: () => this.router.navigate(['/todos']),
      error: () => { this.errorMessage = 'Invalid login or password'; }
    });
  }
}
