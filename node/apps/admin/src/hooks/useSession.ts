import { useQuery } from '@tanstack/react-query';
import { apiFetch, ApiError, businessPaths, type BusinessMeResponse } from '@ting/api';
import { redirectToSignIn } from '../config/auth';

export function useSession(_returnTo = '/admin/items') {
  return useQuery({
    queryKey: ['business', 'me'],
    queryFn: () => apiFetch<BusinessMeResponse>(businessPaths.me),
    retry: false,
  });
}

/** Call after query/mutation errors to redirect on 401. */
export function handleAuthError(err: unknown, returnTo = '/admin/items'): boolean {
  if (err instanceof ApiError && err.status === 401) {
    redirectToSignIn(returnTo);
    return true;
  }
  return false;
}
