import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterOutlet } from '@angular/router';

@Component({
  imports: [CommonModule, RouterOutlet],
  selector: 'app-root',
  styleUrl: './app.css',
  templateUrl: './app.html',
})
export class App {
  private readonly router = inject(Router);

  isAppView(): boolean {
    const url = this.router.url;
    return url.startsWith('/employees');
  }

  navigateTo3D(): void {
    this.router.navigate(['/']);
  }
}
