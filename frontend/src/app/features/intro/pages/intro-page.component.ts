import { Component, HostListener, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';

@Component({
  selector: 'app-intro-page',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './intro-page.component.html',
  styleUrls: ['./intro-page.component.css'],
})
export class IntroPageComponent {
  private readonly router = inject(Router);

  // 3D Tilt rotation angles
  readonly rotX = signal<number>(0);
  readonly rotY = signal<number>(0);
  readonly isTransitioning = signal<boolean>(false);

  @HostListener('mousemove', ['$event'])
  onMouseMove(e: MouseEvent): void {
    if (this.isTransitioning()) return;

    const w = window.innerWidth;
    const h = window.innerHeight;
    const x = e.clientX - w / 2;
    const y = e.clientY - h / 2;

    // Subtle 3D tilt (max ~12 degrees)
    const degX = (y / (h / 2)) * -10;
    const degY = (x / (w / 2)) * 12;

    this.rotX.set(Math.round(degX * 10) / 10);
    this.rotY.set(Math.round(degY * 10) / 10);
  }

  @HostListener('mouseleave')
  onMouseLeave(): void {
    this.rotX.set(0);
    this.rotY.set(0);
  }

  enterDirectory(): void {
    if (this.isTransitioning()) return;
    this.isTransitioning.set(true);

    setTimeout(() => {
      this.router.navigate(['/employees']);
    }, 450);
  }
}
