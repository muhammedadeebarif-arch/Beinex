import { Injectable, inject, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { AdminProfile, LoginPayload, LoginResponse } from '../models/auth.model';

const TOKEN_KEY = 'beinex_auth_token';
const USER_KEY = 'beinex_auth_user';
const INTRO_SEEN_KEY = 'beinex_intro_seen';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly router = inject(Router);

  readonly token = signal<string | null>(this.getStoredToken());
  readonly currentUser = signal<AdminProfile | null>(this.getStoredUser());
  readonly isAuthenticated = computed(() => !!this.token());

  private getStoredToken(): string | null {
    try {
      return localStorage.getItem(TOKEN_KEY);
    } catch {
      return null;
    }
  }

  private getStoredUser(): AdminProfile | null {
    try {
      const raw = localStorage.getItem(USER_KEY);
      return raw ? (JSON.parse(raw) as AdminProfile) : null;
    } catch {
      return null;
    }
  }

  login(payload: LoginPayload): Observable<LoginResponse> {
    return this.http
      .post<LoginResponse>(`${environment.apiUrl}/auth/login`, payload)
      .pipe(
        tap((res) => {
          if (res?.token) {
            localStorage.setItem(TOKEN_KEY, res.token);
            localStorage.setItem(USER_KEY, JSON.stringify(res.admin));
            this.token.set(res.token);
            this.currentUser.set(res.admin);
          }
        }),
      );
  }

  logout(): void {
    try {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(USER_KEY);
    } catch (e) {
      console.warn('Storage error during logout', e);
    }
    this.token.set(null);
    this.currentUser.set(null);
    this.router.navigate(['/login']);
  }

  hasSeenIntro(): boolean {
    try {
      return sessionStorage.getItem(INTRO_SEEN_KEY) === 'true';
    } catch {
      return false;
    }
  }

  markIntroSeen(): void {
    try {
      sessionStorage.setItem(INTRO_SEEN_KEY, 'true');
    } catch (e) {
      console.warn('Session storage error marking intro seen', e);
    }
  }
}
