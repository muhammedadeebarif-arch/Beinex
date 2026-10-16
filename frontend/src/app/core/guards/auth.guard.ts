import { CanActivateFn } from '@angular/router';

// Open access guard - no login gate required per user request
export const authGuard: CanActivateFn = () => {
  return true;
};
