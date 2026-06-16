import { isApiError } from '@ting/api';
import { redirectToSignIn } from '../config/auth';

/** Redirect to sign-in on 401 (queries + mutations). Returns true if redirected. */
export function handleAuthError(err: unknown, returnTo = '/admin/items'): boolean {
  if (isApiError(err) && err.status === 401) {
    redirectToSignIn(returnTo);
    return true;
  }
  return false;
}
