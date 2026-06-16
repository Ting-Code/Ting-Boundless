import { apiFetch, businessPaths, type BusinessMeResponse } from '@ting/api';
import { useApiQuery } from './useApiQuery';

export function useSession(returnTo = '/admin/items') {
  return useApiQuery({
    queryKey: ['business', 'me'],
    queryFn: () => apiFetch<BusinessMeResponse>(businessPaths.me),
    authReturnTo: returnTo,
  });
}

export { handleAuthError } from './authError';
